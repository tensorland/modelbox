package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

const (
	METADATA_BACKEND_MYSQL      = "mysql"
	METADATA_BACKEND_POSTGRES   = "postgres"
	METADATA_BACKEND_INTEGRATED = "integrated"

	BLOB_STORAGE_BACKEND_FS = "filesystem"
	BLOB_STORAGE_BACKEND_S3 = "s3"
)

type MySQLConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"username"`
	Password string `toml:"password"`
	DbName   string `toml:"dbname"`
}

// This is being duplicated from mysql to accomodate specfic configs
// which are not common like ssl and such and offers flexibility for
// the future
type PostgresConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"username"`
	Password string `toml:"password"`
	DbName   string `toml:"dbname"`
}

// Configuration for Timescaledb. Since it's postgres under the hood
// we are adding all the base postgres config options
type TimescaleDbConfig struct {
	PostgresConfig
}

type ServerConfig struct {
	StorageBackend    string                   `toml:"blob_storage"`
	MetadataBackend   string                   `toml:"metadata_storage"`
	MetricsBackend    string                   `toml:"metrics_storage"`
	ListenAddr        string                   `toml:"listen_addr"`
	FileStorage       *FileStorageConfig       `toml:"blob_storage_filesystem"`
	IntegratedStorage *IntegratedStorageConfig `toml:"metadata_storage_integrated"`
	MySQLConfig       *MySQLConfig             `toml:"metadata_storage_mysql"`
	PostgresConfig    *PostgresConfig          `toml:"metadata_storage_postgres"`
	TimescaleDb       *TimescaleDbConfig       `toml:"metrics_storage_timescaledb"`
	PromAddr          string                   `toml:"prometheus_addr"`
}

// Merges empty values of itself with non-empty values of anotherConfig
func (c *ServerConfig) Merge(anotherConfig *ServerConfig) {
	if c.StorageBackend == "" {
		c.StorageBackend = anotherConfig.StorageBackend
	}

	if c.MetadataBackend == "" {
		c.MetadataBackend = anotherConfig.MetadataBackend
	}

	if c.ListenAddr == "" {
		c.ListenAddr = anotherConfig.ListenAddr
	}
	if c.PromAddr == "" {
		c.PromAddr = anotherConfig.PromAddr
	}
}

func (c *ServerConfig) Validate() error {
	return nil
}

type FileStorageConfig struct {
	BaseDir string `toml:"base_dir"`
}

type IntegratedStorageConfig struct {
	Path string `toml:"path"`
}

func defaultServerConfig() *ServerConfig {
	return &ServerConfig{
		StorageBackend:  "file",
		MetadataBackend: "integrated",
		ListenAddr:      ":8080",
		MetricsBackend:  "inmem",
		PromAddr:        ":2112",
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
	if _, err := toml.Decode(string(data), &serverConfig); err != nil {
		return nil, err
	}
	serverConfig.Merge(defaultServerConfig())
	return &serverConfig, nil
}
