package storage

import (
	"fmt"

	"github.com/VividCortex/mysqlerr"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	storageconfig "github.com/tensorland/modelbox/server/storage/config"

	"go.uber.org/zap"
)

type mysqlQueryEngine struct{}

func (*mysqlQueryEngine) isDuplicate(err error) bool {
	if driverErr, ok := err.(*mysql.MySQLError); ok {
		if driverErr.Number == mysqlerr.ER_DUP_ENTRY {
			return true
		}
	}
	return false
}

func (*mysqlQueryEngine) updateMetadata() string {
	return "INSERT INTO metadata (id, parent_id, metadata) VALUES(:id, :parent_id, :metadata) ON DUPLICATE KEY UPDATE id=VALUES(`id`), parent_id=VALUES(`parent_id`), metadata=VALUES(`metadata`)"
}

func (*mysqlQueryEngine) createExperiment() string {
	return "insert into experiments(id, name, owner, namespace, external_id, ml_framework, created_at, updated_at) values(:id, :name, :owner, :namespace, :external_id, :ml_framework, :created_at, :updated_at)"
}

func (*mysqlQueryEngine) createCheckpoint() string {
	return "insert into checkpoints(id, experiment, epoch, metrics, created_at, updated_at) values(:id, :experiment, :epoch, :metrics, :created_at, :updated_at)"
}

func (*mysqlQueryEngine) createModel() string {
	return "insert into models(id, name, owner, namespace, task, description, created_at, updated_at) values(:id, :name, :owner, :namespace, :task, :description, :created_at, :updated_at)"
}

func (*mysqlQueryEngine) listEventsForObject() string {
	return "select id, parent_id, name, source_name, wallclock, metadata from events where parent_id=?"
}

func (*mysqlQueryEngine) createModelVersion() string {
	return `insert into model_versions(id, name, model_id, version, description, ml_framework, unique_tags, created_at, updated_at) values(:id, :name, :model_id, :version, :description, :ml_framework, :unique_tags, :created_at, :updated_at)`
}

func (*mysqlQueryEngine) createAction() string {
	return "insert into actions (id, parent_id, name, arch, params, created_at, updated_at, finished_at) VALUES (:id, :parent_id, :name, :arch, :params, :created_at, :updated_at, :finished_at)"
}

func (*mysqlQueryEngine) createActionEval() string {
	return "insert into action_evals (id, parent_id, parent_type, eval_type, created_at, processed_at) VALUES (:id, :parent_id, :parent_type, :eval_type, :created_at, :processed_at)"
}

func (*mysqlQueryEngine) blobMultiWrite() string {
	return "insert into blobs(id, parent_id, metadata) VALUES "
}

func (*mysqlQueryEngine) actionInstances() string {
	return "select id, action_id, attempt, status, outcome, outcome_reason, created_at, updated_at, finished_at from action_instances where action_id=?"
}

func (*mysqlQueryEngine) getActionEval() string {
	return "select id, parent_id, parent_type, eval_type, created_at, processed_at from action_evals where id=?"
}

func (*mysqlQueryEngine) getActionInstance() string {
	return "select id, action_id, attempt, status, outcome, outcome_reason, created_at, updated_at, finished_at from action_instances where id=?"
}

type MySqlStorage struct {
	*SQLStorage
	*MYSQLDriverUtils
	db     *sqlx.DB
	config *storageconfig.MySqlStorageConfig

	logger *zap.Logger
}

func NewMySqlStorage(config *storageconfig.MySqlStorageConfig, logger *zap.Logger) (*MySqlStorage, error) {
	db, err := sqlx.Open("mysql", config.DataSource())
	if err != nil {
		return nil, fmt.Errorf("unable to connect to mysql %v", err)
	}
	sqlStorage := NewSQLStorage(db, &mysqlQueryEngine{}, logger)
	util := &MYSQLDriverUtils{Config: config, Db: db, Logger: logger}
	return &MySqlStorage{SQLStorage: sqlStorage, MYSQLDriverUtils: util, db: db, config: config, logger: logger}, nil
}

func (m *MySqlStorage) Ping() error {
	return m.db.Ping()
}

func (m *MySqlStorage) Close() error {
	return m.db.Close()
}

func (e *MySqlStorage) Backend() *BackendInfo {
	return &BackendInfo{Name: "mysql"}
}
