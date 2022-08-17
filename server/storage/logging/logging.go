package logging

import (
	"context"

	"github.com/diptanu/modelbox/server/config"
	"go.uber.org/zap"
)

type FloatLog struct {
	Value     float32
	Step      uint64
	WallClock uint64
}
type ExperimentLogger interface {
	LogFloats(ctx context.Context, parentId, key string, value *FloatLog) error

	GetFloatLogs(ctx context.Context, parentId string) (map[string][]*FloatLog, error)

	Backend() string
}

func NewExperimentLogger(serverConfig *config.ServerConfig, logger *zap.Logger) (ExperimentLogger, error) {
	if serverConfig.MetadataBackend == config.METRICS_STORAGE_TS {
		return NewTimescaleDbLogger(&TimescaleDbConfig{
			Host:     serverConfig.TimescaleDb.Host,
			Port:     serverConfig.MySQLConfig.Port,
			UserName: serverConfig.TimescaleDb.User,
			Password: serverConfig.TimescaleDb.Password,
			DbName:   serverConfig.TimescaleDb.DbName,
		}, logger)
	}
	return NewInMemoryExperimentLogger()
}
