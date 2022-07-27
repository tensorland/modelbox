package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModelIdUniqueNess(t *testing.T) {
	m1 := NewModel(MODEL_NAME, OWNER, NAMESPACE, TASK, "text to text transalte", nil)
	m2 := NewModel(MODEL_NAME, OWNER, "namespace-x", TASK, "text to text transalte", nil)
	assert.NotEqual(t, m1.Id, m2.Id)
}
