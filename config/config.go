package config

import (
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
}

type Config struct {
	Rules []ForwardRule `yaml:"rules"`
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

	return &config, nil
}

func SaveConfig(filename string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
