package logging

import (
	"context"
	"fmt"

	"github.com/tensorland/modelbox/server/config"
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
	if serverConfig.MetricsBackend == config.METRICS_STORAGE_TS {
		return NewTimescaleDbLogger(&TimescaleDbConfig{
			Host:     serverConfig.TimescaleDb.Host,
			Port:     serverConfig.TimescaleDb.Port,
			UserName: serverConfig.TimescaleDb.User,
			Password: serverConfig.TimescaleDb.Password,
			DbName:   serverConfig.TimescaleDb.DbName,
		}, logger)
	} else if serverConfig.MetricsBackend == config.METRICS_STORAGE_INMEMORY {
		return NewInMemoryExperimentLogger()
	}
	return nil, fmt.Errorf("unable to create experiment logger for driver: %v", serverConfig.MetricsBackend)
}
