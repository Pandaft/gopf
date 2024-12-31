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

	"github.com/charmbracelet/lipgloss"
)

const (
	version           = "v0.1.1"
	defaultConfigFile = "gopf.yaml"
)

// UI样式定义
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginLeft(2)

	versionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
)

// 初始化空配置文件
func initEmptyConfig(filename string) (*config.Config, error) {
	cfg := &config.Config{
		Rules: make([]config.ForwardRule, 0),
	}

	if err := config.SaveConfig(filename, cfg); err != nil {
		return nil, fmt.Errorf("创建配置文件失败: %v", err)
	}

	return cfg, nil
}

// 加载配置文件
func loadConfig(filename string) (*config.Config, error) {
	cfg, err := config.LoadConfig(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("配置文件 %s 不存在，创建新的配置文件...", filename)
			if _, err = initEmptyConfig(filename); err != nil {
				return nil, fmt.Errorf("初始化配置文件失败: %v", err)
			}
			// 重新加载配置文件
			return config.LoadConfig(filename)
		}
		return nil, fmt.Errorf("加载配置文件失败: %v", err)
	}
	return cfg, nil
}

// 启动所有转发规则
func startForwarders(cfg *config.Config) map[string]*forwarder.Forwarder {
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
	return forwarders
}

// 设置信号处理
func setupSignalHandler(forwarders map[string]*forwarder.Forwarder) {
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
}

// 显示版本信息
func printVersion() {
	title := titleStyle.Render("GOPF")
	ver := versionStyle.Render(version)
	fmt.Printf("%s %s\n", title, ver)
}

func main() {
	printVersion()

	// 解析命令行参数
	configFile := flag.String("config", defaultConfigFile, "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := loadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	// 启动转发器
	forwarders := startForwarders(cfg)

	// 设置信号处理
	setupSignalHandler(forwarders)

	// 启动UI
	if err := ui.StartUI(cfg, forwarders); err != nil {
		log.Fatalf("UI启动失败: %v", err)
	}
}
