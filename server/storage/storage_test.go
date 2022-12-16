package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestModelIdUniqueNess(t *testing.T) {
	m1 := NewModel(MODEL_NAME, OWNER, NAMESPACE, TASK, "text to text transalte")
	m2 := NewModel(MODEL_NAME, OWNER, "namespace-x", TASK, "text to text transalte")
	assert.NotEqual(t, m1.Id, m2.Id)
}

func TestEventIdUniquess(t *testing.T) {
	timeNow := time.Now()
	e1 := NewEvent("foo", "trainer", "data_download_start", timeNow, map[string]*structpb.Value{})
	e2 := NewEvent("foo", "trainer", "data_download_finish", timeNow, map[string]*structpb.Value{})
	assert.NotEqual(t, e1.Id, e2.Id)

	// Same events but different time
	e3 := NewEvent("foo", "trainer", "data_download_start", timeNow.Add(2*time.Second), map[string]*structpb.Value{})
	e4 := NewEvent("foo", "trainer", "data_download_finish", timeNow.Add(2*time.Second), map[string]*structpb.Value{})
	assert.NotEqual(t, e1.Id, e3.Id)
	assert.NotEqual(t, e2.Id, e4.Id)
}

func TestActionIdUniqueNess(t *testing.T) {
	act := NewAction("quantize1", "x86", "parent1", NewTrigger("", TriggerTypeJs), createMetadata())

	act1 := NewAction("quantize", "x86", "parent1", NewTrigger("", TriggerTypeJs), createMetadata())

	assert.NotEqual(t, act, act1)
}

func createMetadata() map[string]*structpb.Value {
	metaVal, _ := structpb.NewValue(map[string]interface{}{"/foo": 5})
	return map[string]*structpb.Value{"foo": metaVal}
}
