package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Language string

const (
	Chinese Language = "zh"
	English Language = "en"
)

type ForwardRule struct {
	Name        string `yaml:"name"`
	LocalPort   int    `yaml:"local_port"`
	RemoteHost  string `yaml:"remote_host"`
	RemotePort  int    `yaml:"remote_port"`
	BytesSent   uint64 `yaml:"-"`
	BytesRecv   uint64 `yaml:"-"`
	Connections uint64 `yaml:"-"`
	Status      string `yaml:"-"`
	Error       string `yaml:"-"`
	IsRunning   bool   `yaml:"-"`
	LastActive  int64  `yaml:"-"`
}

type Config struct {
	Rules      []ForwardRule `yaml:"rules"`
	configPath string        `yaml:"-"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	config.configPath = filename
	return &config, nil
}

func (c *Config) Save() error {
	if c.configPath == "" {
		return fmt.Errorf("配置文件路径未设置")
	}
	return SaveConfig(c.configPath, c)
}

func (c *Config) AddRule(rule ForwardRule) error {
	// 检查端口是否已存在
	for _, r := range c.Rules {
		if r.LocalPort == rule.LocalPort {
			return fmt.Errorf("本地端口 %d 已被使用", rule.LocalPort)
		}
	}

	c.Rules = append(c.Rules, rule)
	return c.Save()
}

func (c *Config) UpdateRule(index int, rule ForwardRule) error {
	if index < 0 || index >= len(c.Rules) {
		return fmt.Errorf("规则索引越界")
	}

	// 检查端口是否已被其他规则使用
	for i, r := range c.Rules {
		if i != index && r.LocalPort == rule.LocalPort {
			return fmt.Errorf("本地端口 %d 已被规则 '%s' 使用", rule.LocalPort, r.Name)
		}
	}

	c.Rules[index] = rule
	return c.Save()
}

func (c *Config) DeleteRule(index int) error {
	if index < 0 || index >= len(c.Rules) {
		return fmt.Errorf("规则索引越界")
	}

	// 删除指定索引的规则
	c.Rules = append(c.Rules[:index], c.Rules[index+1:]...)
	return c.Save()
}

func SaveConfig(filename string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
