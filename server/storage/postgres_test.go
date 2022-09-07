package storage

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

func env(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type PostgresTestSuite struct {
	suite.Suite
	StorageInterfaceTestSuite
	pqStorage *PostgresStorage
}

func (s *PostgresTestSuite) SetupSuite() {
	if err := s.pqStorage.CreateSchema("schemas/postgres/schema_ver_0.sql"); err != nil {
		s.T().Fatalf("unable to load postgres schema: %v", err)
	}
}

func (s *PostgresTestSuite) TearDownSuite() {
	if err := s.pqStorage.DropDb(); err != nil {
		s.T().Fatalf("unable to drop db: %v", err)
	}
}

func TestPostgresTestSuite(t *testing.T) {
	logger, _ := zap.NewProduction()
	port, _ = strconv.ParseUint(env("POSTGRES_TEST_PORT", "5432"), 10, 64)
	config := PostgresConfig{
		Host:     env("POSTGRES_TEST_HOST", "172.20.0.5"),
		Port:     int(port),
		UserName: env("POSTGRES_TEST_USER", "postgres"),
		Password: env("POSTGRES_TEST_PASS", "foo"),
		DbName:   env("POSTGRES_TEST_DBNAME", "gotest"),
	}
	postgres, err := NewPostgresStorage(&config, logger)
	if err != nil {
		t.Fatalf("unable to create mysql: %v", err)
	}
	suite.Run(t, &PostgresTestSuite{
		StorageInterfaceTestSuite: StorageInterfaceTestSuite{
			t:         t,
			storageIf: postgres,
		},
		pqStorage: postgres,
	})
}
