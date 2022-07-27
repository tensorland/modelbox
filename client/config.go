package client

import (
	"io/ioutil"
	"net"

	"github.com/BurntSushi/toml"
)

type ClientConfig struct {
	ServerAddr string `toml:"server_addr"`
}

func NewClientConfig(configPath string) (*ClientConfig, error) {
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config ClientConfig
	if _, err := toml.Decode(string(bytes), &config); err != nil {
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
