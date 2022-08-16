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

func NewExperimentLogger(config *config.ServerConfig, logger *zap.Logger) (ExperimentLogger, error) {
	if config.MetricsBackend == "timescaledb" {
		return NewTimescaleDbLogger(&TimescaleDbConfig{}, logger)
	}
	return NewInMemoryExperimentLogger(), nil
}
