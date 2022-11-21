package scheduler

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tensorland/modelbox/server/storage"
	storageconfig "github.com/tensorland/modelbox/server/storage/config"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const SQLITE3_FILE = "/tmp/test_modelbox_scheudler_sqlite3.dat"

type SchedulerTestSuite struct {
	suite.Suite

	actionScheduler ActionScheduler

	storageIf storage.MetadataStorage
}

func (s *SchedulerTestSuite) SetupSuite() {
	if err := s.storageIf.CreateSchema("../storage/schemas/sqlite3/"); err != nil {
		s.T().Fatalf("unable to load sqlite schema: %v", err)
	}
}

func (s *SchedulerTestSuite) TearDownSuite() {
	os.Remove(SQLITE3_FILE)
}

func (s *SchedulerTestSuite) TestCreateNewAction() {
	ctx := context.Background()
	act := storage.NewAction("quantize", "x86", "parent1", s.createMetadata())
	err := s.storageIf.CreateAction(ctx, act)
	assert.Nil(s.T(), err)

	err = s.actionScheduler.runScheduler()
	require.Nil(s.T(), err)
	actionState, err := s.storageIf.GetAction(ctx, act.Id)
	require.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(actionState.Instances))
}

func (s *SchedulerTestSuite) TestFinishAction() {
}

func (s *SchedulerTestSuite) TestFailAction() {
}

func (s *SchedulerTestSuite) createMetadata() map[string]*structpb.Value {
	metaVal, _ := structpb.NewValue(map[string]interface{}{"/foo": 5})
	return map[string]*structpb.Value{"foo": metaVal}
}

func TestSchedulerTestSuite(t *testing.T) {
	logger, _ := zap.NewProduction()
	config := storageconfig.Sqlite3Config{
		File: SQLITE3_FILE,
	}
	sqliteStorage, err := storage.NewSqlite3Storage(&config, logger)
	if err != nil {
		t.Fatalf("unable to create sqlite3: %v", err)
	}
	scheduler := NewActionScheduler(sqliteStorage, 5*time.Second, logger)
	suite.Run(t, &SchedulerTestSuite{
		actionScheduler: *scheduler,
		storageIf:       sqliteStorage,
	})
}
