package scheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tensorland/modelbox/server/storage"
	"go.uber.org/zap"
)

func TestEvaluateTrigger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	triggerEvaluator := NewTriggerEvaluator(logger)
	changeEvent := storage.NewChangeEvent("foo", 10, storage.EventTypeExperimentCreated, storage.EventObjectTypeExperiment, "", storage.NewExperiment("foo", "bar", "lol", "", storage.Pytorch))
	action, err := triggerEvaluator.GetAction(changeEvent, storage.NewTrigger(`action = "foo";action;`, storage.TriggerTypeJs))
	require.Nil(t, err)
	assert.Equal(t, "foo", action)
}
