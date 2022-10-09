package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	METADATA_BACKEND_MYSQL     = "mysql"
	METADATA_BACKEND_POSTGRES  = "postgres"
	METADATA_BACKEND_EPHEMERAL = "ephemeral"

	BLOB_STORAGE_BACKEND_FS = "filesystem"
	BLOB_STORAGE_BACKEND_S3 = "s3"

	METRICS_STORAGE_TS       = "timescaledb"
	METRICS_STORAGE_INMEMORY = "inmemory"
)

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"username"`
	Password string `yaml:"password"`
	DbName   string `yaml:"dbname"`
}

// This is being duplicated from mysql to accomodate specfic configs
// which are not common like ssl and such and offers flexibility for
// the future
type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"username"`
	Password string `yaml:"password"`
	DbName   string `yaml:"dbname"`
}

// Configuration to use SQL datastores for cluster membership
type SQLClusterMembership struct {
	// Pings the database to renew lease
	LeaseInterval time.Duration `yaml:"lease_interval"`

	StaleHeartbeatDuraion time.Duration `yaml:"stale_heartbeat_duration"`
}

// Represents hosts participating in a static cluster
type ClusterMember struct {
	Id       string `yaml:"id"`
	HostName string `yaml:"host_name"`
	RPCAddr  string `yaml:"rpc_addr"`
	HttpAddr string `yaml:"http_addr"`
}

type StaticClusterMembership struct {
	Members []*ClusterMember `yaml:"members"`
}

// Configuration for Timescaledb. Since it's postgres under the hood
// we are adding all the base postgres config options
type TimescaleDbConfig struct {
	PostgresConfig `yaml:",inline"`
}

type ServerConfig struct {
	ArtifactStorageBackend   string                   `yaml:"artifact_storage"`
	MetadataBackend          string                   `yaml:"metadata_storage"`
	MetricsBackend           string                   `yaml:"metrics_storage"`
	GrpcListenAddr           string                   `yaml:"grpc_listen_addr"`
	HttpListenAddr           string                   `yaml:"http_listen_addr"`
	FileStorage              *FileStorageConfig       `yaml:"artifact_storage_filesystem"`
	S3Storage                *S3StorageConfig         `yaml:"artifact_storage_s3"`
	IntegratedStorage        *IntegratedStorageConfig `yaml:"metadata_storage_integrated"`
	MySQLConfig              *MySQLConfig             `yaml:"metadata_storage_mysql"`
	PostgresConfig           *PostgresConfig          `yaml:"metadata_storage_postgres"`
	TimescaleDb              *TimescaleDbConfig       `yaml:"metrics_storage_timescaledb"`
	PromAddr                 string                   `yaml:"prometheus_addr"`
	ClusterMembershipBackend string                   `yaml:"cluster_membership"`
	SQLClusterMembership     *SQLClusterMembership    `yaml:"sql_cluster_membership"`
	StaticClusterMembership  *StaticClusterMembership `yaml:"static_cluster_membership"`
}

// Merges empty values of itself with non-empty values of anotherConfig
func (c *ServerConfig) Merge(anotherConfig *ServerConfig) {
	if c.ArtifactStorageBackend == "" {
		c.ArtifactStorageBackend = anotherConfig.ArtifactStorageBackend
	}

	if c.MetadataBackend == "" {
		c.MetadataBackend = anotherConfig.MetadataBackend
	}

	if c.MetricsBackend == "" {
		c.MetricsBackend = anotherConfig.MetricsBackend
	}

	if c.GrpcListenAddr == "" {
		c.GrpcListenAddr = anotherConfig.GrpcListenAddr
	}
	if c.HttpListenAddr == "" {
		c.HttpListenAddr = anotherConfig.HttpListenAddr
	}
	if c.PromAddr == "" {
		c.PromAddr = anotherConfig.PromAddr
	}
}

func (c *ServerConfig) Validate() error {
	return nil
}

type FileStorageConfig struct {
	BaseDir string `yaml:"base_dir"`
}

type S3StorageConfig struct {
	Region string `yaml:"region"`
	Bucket string `yaml:"bucket"`
}

type IntegratedStorageConfig struct {
	Path string `yaml:"path"`
}

func defaultServerConfig() *ServerConfig {
	return &ServerConfig{
		ArtifactStorageBackend:   "filesystem",
		MetadataBackend:          "ephemeral",
		GrpcListenAddr:           ":8080",
		HttpListenAddr:           ":8085",
		MetricsBackend:           "inmemory",
		PromAddr:                 ":2112",
		ClusterMembershipBackend: "static",
		StaticClusterMembership: &StaticClusterMembership{
			Members: []*ClusterMember{
				{
					Id:      "localhost",
					RPCAddr: ":8080",
				},
			},
		},
	}
}

type LoggingConfig struct {
	LogLevel          string
	LogJson           bool
	LogFile           string
	LogRotateDuration string
	LogRotateBytes    uint64
	LogRotateMaxFiles uint32
}

func NewServerConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read server config: %v", err)
	}
	var serverConfig ServerConfig
	if err := yaml.Unmarshal(data, &serverConfig); err != nil {
		return nil, err
	}
	serverConfig.Merge(defaultServerConfig())
	return &serverConfig, nil
}
