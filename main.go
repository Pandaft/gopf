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
			rule.Error = err.Error()
			log.Printf("警告: 端口转发启动失败 [%s]: %v\n", rule.Name, err)
			continue
		}
		rule.IsRunning = true
		forwarders[rule.Name] = f
	}

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n正在关闭所有端口转发...")
		for _, f := range forwarders {
			f.Stop()
		}
		os.Exit(0)
	}()

	// 启动UI
	if err := ui.StartUI(cfg, forwarders); err != nil {
		log.Fatalf("UI启动失败: %v", err)
	}
}
