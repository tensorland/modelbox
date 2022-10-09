package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerConfig(t *testing.T) {
	config, err := NewServerConfig("./../../cmd/modelbox/assets/modelbox_server.yaml")
	assert.Nil(t, err)
	assert.Equal(t, ":8086", config.GrpcListenAddr)
	assert.Equal(t, ":8081", config.HttpListenAddr)
	assert.Equal(t, "filesystem", config.ArtifactStorageBackend)
	assert.Equal(t, "ephemeral", config.MetadataBackend)
	assert.Equal(t, "inmemory", config.MetricsBackend)

	assert.Equal(t, "/tmp/modelboxblobs", config.FileStorage.BaseDir)

	assert.Equal(t, "/tmp/modelbox.dat", config.IntegratedStorage.Path)
}

func TestS3Config(t *testing.T) {
	config, err := NewServerConfig("./../../cmd/modelbox/assets/modelbox_server.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, config.S3Storage)
	assert.Equal(t, "us-east-1", config.S3Storage.Region)
	assert.Equal(t, "modelbox-artifacts", config.S3Storage.Bucket)
}

func TestMySQLConfig(t *testing.T) {
	config, err := NewServerConfig("./../../cmd/modelbox/assets/modelbox_server.yaml")
	assert.Nil(t, err)
	assert.Equal(t, "172.17.0.2", config.MySQLConfig.Host)
	assert.Equal(t, 3306, config.MySQLConfig.Port)
	assert.Equal(t, "root", config.MySQLConfig.User)
	assert.Equal(t, "foo", config.MySQLConfig.Password)
	assert.Equal(t, "modelbox", config.MySQLConfig.DbName)
}

func TestPostgresConfigTest(t *testing.T) {
	config, err := NewServerConfig("./../../cmd/modelbox/assets/modelbox_server.yaml")
	assert.Nil(t, err)
	assert.Equal(t, "172.17.0.3", config.PostgresConfig.Host)
	assert.Equal(t, 5432, config.PostgresConfig.Port)
	assert.Equal(t, "postgres", config.PostgresConfig.User)
	assert.Equal(t, "foo", config.PostgresConfig.Password)
	assert.Equal(t, "modelbox", config.PostgresConfig.DbName)
}

func TestTimescaledbConfigTest(t *testing.T) {
	config, err := NewServerConfig("./../../cmd/modelbox/assets/modelbox_server.yaml")
	assert.Nil(t, err)
	assert.Equal(t, "172.17.0.4", config.TimescaleDb.Host)
	assert.Equal(t, 5432, config.TimescaleDb.Port)
	assert.Equal(t, "postgres", config.TimescaleDb.User)
	assert.Equal(t, "foo", config.TimescaleDb.Password)
	assert.Equal(t, "modelbox_metrics", config.TimescaleDb.DbName)
}

func TestStaticClusterMembershipTest(t *testing.T) {
	config, err := NewServerConfig("./../../cmd/modelbox/assets/modelbox_server.yaml")
	assert.Nil(t, err)
	assert.Equal(t, "static", config.ClusterMembershipBackend)

	assert.Equal(t, 1, len(config.StaticClusterMembership.Members))
	member := config.StaticClusterMembership.Members[0]
	assert.Equal(t, "host1", member.Id)
	assert.Equal(t, "host1", member.HostName)
	assert.Equal(t, "localhost:8086", member.RPCAddr)
	assert.Equal(t, "localhost:8081", member.HttpAddr)
}

func TestSQLClusterMembership(t *testing.T) {
	config, err := NewServerConfig("./../../cmd/modelbox/assets/modelbox_server.yaml")
	assert.Nil(t, err)

	assert.Equal(t, 5*time.Second, config.SQLClusterMembership.LeaseInterval)
	assert.Equal(t, 30*time.Second, config.SQLClusterMembership.StaleHeartbeatDuraion)
}
