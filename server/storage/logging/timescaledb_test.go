package logging

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

func env(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type TimescaleDbTestSuite struct {
	suite.Suite
	ts *TimescaleDbLogger
}

func (s *TimescaleDbTestSuite) SetupSuite() {
	if err := s.ts.CreateSchema("../schemas/timescaledb/schema_ver_0.sql"); err != nil {
		s.T().Fatalf("unable to load timescaledb schema: %v", err)
	}
}

func (s *TimescaleDbTestSuite) TearDownSuite() {
	if err := s.ts.Close(); err != nil {
		s.T().Fatalf("unable to close db: %v", err)
	}
	if err := s.ts.DropDb(); err != nil {
		s.T().Fatalf("unable to drop db: %v", err)
	}
}

func (s *TimescaleDbTestSuite) TestLogFloat() {
	log1 := &FloatLog{Value: 0.9, Step: 1, WallClock: 10000}
	log2 := &FloatLog{Value: 0.95, Step: 3, WallClock: 15000}
	ctx := context.Background()
	err := s.ts.LogFloats(ctx, "parent1", "val_accu", log1)
	assert.Nil(s.T(), err)
	err = s.ts.LogFloats(ctx, "parent1", "val_accu", log2)
	assert.Nil(s.T(), err)

	// Retrieve the logs
	logs, err := s.ts.GetFloatLogs(ctx, "parent1")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(logs))
	assert.Equal(s.T(), 2, len(logs["val_accu"]))
}

func TestPostgresTestSuite(t *testing.T) {
	logger, _ := zap.NewProduction()
	port, _ := strconv.ParseUint(env("POSTGRES_TEST_PORT", "5432"), 10, 64)
	config := TimescaleDbConfig{
		Host:     env("POSTGRES_TEST_HOST", "172.20.0.7"),
		Port:     int(port),
		UserName: env("POSTGRES_TEST_USER", "postgres"),
		Password: env("POSTGRES_TEST_PASS", "foo"),
		DbName:   env("POSTGRES_TEST_DBNAME", "gotest"),
	}
	ts, err := NewTimescaleDbLogger(&config, logger)
	if err != nil {
		t.Fatalf("unable to create timescaledb: %v", err)
	}
	suite.Run(t, &TimescaleDbTestSuite{
		ts: ts,
	})
}
