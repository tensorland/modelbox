package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	storageconfig "github.com/tensorland/modelbox/server/storage/config"
	"go.uber.org/zap"
)

const SQLITE3_FILE = "/tmp/test_modelbox_sqlite3.dat"

type SqliteTestSuite struct {
	suite.Suite
	StorageInterfaceTestSuite
	sqliteStorage *Sqlite3Storage
}

func (s *SqliteTestSuite) SetupSuite() {
	if err := s.sqliteStorage.CreateSchema("schemas/sqlite3/"); err != nil {
		s.T().Fatalf("unable to load sqlite schema: %v", err)
	}
}

func (s *SqliteTestSuite) TearDownSuite() {
	os.Remove(SQLITE3_FILE)
}

func TestSqliteTestSuite(t *testing.T) {
	logger, _ := zap.NewProduction()
	config := storageconfig.Sqlite3Config{
		File: SQLITE3_FILE,
	}
	sqliteStorage, err := NewSqlite3Storage(&config, logger)
	if err != nil {
		t.Fatalf("unable to create sqlite3: %v", err)
	}
	suite.Run(t, &SqliteTestSuite{
		StorageInterfaceTestSuite: StorageInterfaceTestSuite{
			t:         t,
			storageIf: sqliteStorage,
		},
		sqliteStorage: sqliteStorage,
	})
}
