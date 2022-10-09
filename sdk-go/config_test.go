package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientConfig(t *testing.T) {
	config, err := NewClientConfig("../cmd/modelbox/assets/modelbox_client.yaml")
	assert.Nil(t, err)
	assert.Equal(t, ":8085", config.ServerAddr)
}
