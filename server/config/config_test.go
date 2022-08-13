package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerConfig(t *testing.T) {
	config, err := NewServerConfig("./../../cmd/modelbox/assets/modelbox_server.toml")
	assert.Nil(t, err)
	assert.Equal(t, ":8085", config.ListenAddr)
	assert.Equal(t, "filesystem", config.StorageBackend)
	assert.Equal(t, "integrated", config.MetadataBackend)

	assert.Equal(t, "/tmp/modelboxblobs", config.FileStorage.BaseDir)

	assert.Equal(t, "/tmp/modelbox.dat", config.IntegratedStorage.Path)
}

func TestMySQLConfig(t *testing.T) {
	config, err := NewServerConfig("./../../cmd/modelbox/assets/modelbox_server.toml")
	assert.Nil(t, err)
	assert.Equal(t, "172.17.0.2", config.MySQLConfig.Host)
	assert.Equal(t, 3306, config.MySQLConfig.Port)
	assert.Equal(t, "root", config.MySQLConfig.User)
	assert.Equal(t, "foo", config.MySQLConfig.Password)
	assert.Equal(t, "modelbox", config.MySQLConfig.DbName)
}

func TestPostgresConfigTest(t *testing.T) {
	config, err := NewServerConfig("./../../cmd/modelbox/assets/modelbox_server.toml")
	assert.Nil(t, err)
	assert.Equal(t, "172.17.0.3", config.PostgresConfig.Host)
	assert.Equal(t, 5432, config.PostgresConfig.Port)
	assert.Equal(t, "postgres", config.PostgresConfig.User)
	assert.Equal(t, "foo", config.PostgresConfig.Password)
	assert.Equal(t, "modelbox", config.PostgresConfig.DbName)
}
