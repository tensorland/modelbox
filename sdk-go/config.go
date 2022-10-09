package client

import (
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

type ClientConfig struct {
	ServerAddr string `yaml:"server_addr"`
}

func NewClientConfig(configPath string) (*ClientConfig, error) {
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config ClientConfig
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}
	if err := config.validate(); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *ClientConfig) validate() error {
	if _, _, err := net.SplitHostPort(c.ServerAddr); err != nil {
		return err
	}
	return nil
}
