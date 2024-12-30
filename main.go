package main

import (
	"flag"
	"fmt"
	"gopf/config"
	"gopf/forwarder"
	"gopf/ui"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configFile := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	forwarders := make(map[string]*forwarder.Forwarder)
	for i := range cfg.Rules {
		rule := &cfg.Rules[i]
		f := forwarder.NewForwarder(rule)
		if err := f.Start(); err != nil {
			log.Fatalf("启动转发器失败 [%s]: %v", rule.Name, err)
		}
		forwarders[rule.Name] = f
	}

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n正在关闭所有转发器...")
		for _, f := range forwarders {
			f.Stop()
		}
		os.Exit(0)
	}()

	// 启动UI
	if err := ui.StartUI(cfg.Rules, forwarders); err != nil {
		log.Fatalf("UI启动失败: %v", err)
	}
}
