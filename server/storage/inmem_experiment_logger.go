package storage

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type InMemoryExperimentLogger struct {
	floatLogs     map[string][]*FloatLog
	floatLogsLock sync.RWMutex
}

func NewInMemoryExperimentLogger() *InMemoryExperimentLogger {
	return &InMemoryExperimentLogger{
		floatLogs: make(map[string][]*FloatLog),
	}
}

func (i *InMemoryExperimentLogger) LogFloats(ctx context.Context, parentId string, key string, value *FloatLog) error {
	i.floatLogsLock.Lock()
	defer i.floatLogsLock.Unlock()
	compoundKey := fmt.Sprintf("%s-%s", parentId, key)
	i.floatLogs[compoundKey] = append(i.floatLogs[compoundKey], value)
	return nil
}

func (i *InMemoryExperimentLogger) GetFloatLogs(ctx context.Context, parentId string) (map[string][]*FloatLog, error) {
	i.floatLogsLock.RLock()
	defer i.floatLogsLock.RUnlock()
	logs := make(map[string][]*FloatLog)
	for k := range i.floatLogs {
		if strings.HasPrefix(k, parentId) {
			prefix := fmt.Sprintf("%s-", parentId)
			key := strings.TrimPrefix(k, prefix)
			logs[key] = i.floatLogs[k]
		}
	}
	return logs, nil
}
