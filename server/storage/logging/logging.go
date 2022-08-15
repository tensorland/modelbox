package logging

import "context"

type FloatLog struct {
	Value     float32
	Step      uint64
	WallClock uint64
}
type ExperimentLogger interface {
	LogFloats(ctx context.Context, parentId string, key string, value *FloatLog) error

	GetFloatLogs(ctx context.Context, parentId string) (map[string][]*FloatLog, error)
}
