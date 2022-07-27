package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

const (
	STORAGE_HANDLE = "/tmp/ephemeral_storage_test"
)

type EphemeralStorageTestSuite struct {
	suite.Suite
	StorageInterfaceTestSuite
	storage *EphemeralStorage
}

func TestEphemeralStoreTestSuite(t *testing.T) {
	logger, _ := zap.NewProduction()
	s, err := NewEphemeralStorage(STORAGE_HANDLE, logger)
	if err != nil {
		t.Fatalf("couldn't create storage")
	}
	defer s.Close()
	os.Remove(STORAGE_HANDLE)
	suite.Run(t, &EphemeralStorageTestSuite{
		StorageInterfaceTestSuite: StorageInterfaceTestSuite{
			t:         t,
			storageIf: s,
		},
		storage: s,
	})
}
