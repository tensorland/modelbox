package scheduler

import (
	"fmt"

	"github.com/robertkrimen/otto"
	"github.com/tensorland/modelbox/server/storage"
	"go.uber.org/zap"
)

type TriggerEvaluator struct {
	vm     *otto.Otto
	logger *zap.Logger
}

func NewTriggerEvaluator(logger *zap.Logger) *TriggerEvaluator {
	vm := otto.New()
	return &TriggerEvaluator{vm: vm, logger: logger}
}

func (e *TriggerEvaluator) GetAction(event *storage.ChangeEvent, t *storage.Trigger) (string, error) {
	if event == nil {
		return "", fmt.Errorf("no changevent to evaluate trigger")
	}
	if err := e.vm.Set("changeEvent", *event); err != nil {
		return "", fmt.Errorf("unable to set change event: %v", err)
	}
	val, err := e.vm.Run(t.Payload)
	if err != nil {
		return "", fmt.Errorf("unable to evaluate trigger: %v", err)
	}
	action, err := val.ToString()
	if err != nil {
		return "", fmt.Errorf("unable to get action name from trigger: %v", err)
	}
	return action, nil
}
