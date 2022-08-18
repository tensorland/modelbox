package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientConfig(t *testing.T) {
	config, err := NewClientConfig("assets/modelbox_client.toml")
	assert.Nil(t, err)
	assert.Equal(t, ":8085", config.ServerAddr)
}
