package storage

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/lib/pq"
	storageconfig "github.com/tensorland/modelbox/server/storage/config"
	"go.uber.org/zap"
)

const (
	POSTGRES_DRIVER = "postgres"
)

type postgresQueryEngine struct{}

func (*postgresQueryEngine) isDuplicate(err error) bool {
	if driverErr, ok := err.(*pq.Error); ok {
		if driverErr.Code == "23505" {
			return true
		}
	}
	return false
}

func (*postgresQueryEngine) updateMetadata() string {
	return "INSERT INTO metadata (id, parent_id, metadata) VALUES(:id, :parent_id, :metadata) ON CONFLICT (id) DO UPDATE SET updated_at  = EXCLUDED.updated_at, metadata = EXCLUDED.metadata "
}

func (*postgresQueryEngine) createExperiment() string {
	return "insert into experiments(id, name, owner, namespace, external_id, ml_framework, created_at, updated_at) values(:id, :name, :owner, :namespace, :external_id, :ml_framework, :created_at, :updated_at) ON CONFLICT (id) DO UPDATE SET updated_at = EXCLUDED.updated_at"
}

func (*postgresQueryEngine) createCheckpoint() string {
	return "insert into checkpoints(id, experiment, epoch, metrics, created_at, updated_at) values(:id, :experiment, :epoch, :metrics, :created_at, :updated_at) ON CONFLICT (id) DO UPDATE SET updated_at = EXCLUDED.updated_at"
}

func (*postgresQueryEngine) createModel() string {
	return "insert into models(id, name, owner, namespace, task, description, created_at, updated_at) values(:id, :name, :owner, :namespace, :task, :description, :created_at, :updated_at) ON CONFLICT (id) DO UPDATE SET updated_at = EXCLUDED.updated_at"
}

func (*postgresQueryEngine) listEventsForObject() string {
	return "select id, parent_id, name, source_name, wallclock, metadata from events where parent_id=$1"
}

func (*postgresQueryEngine) createModelVersion() string {
	return `insert into model_versions(id, name, model_id, version, description, ml_framework, unique_tags, created_at, updated_at) values(:id, :name, :model_id, :version, :description, :ml_framework, :unique_tags, :created_at, :updated_at)`
}

func (*postgresQueryEngine) createAction() string {
	return "insert into actions (id, parent_id, name, arch, trigger_predicate, params, created_at, updated_at, finished_at) VALUES (:id, :parent_id, :name, :arch, :trigger_predicate, :params, :created_at, :updated_at, :finished_at)"
}

func (*postgresQueryEngine) blobMultiWrite() string {
	return "insert into blobs(id, parent_id, metadata) VALUES "
}

func (*postgresQueryEngine) createActionEval() string {
	return "insert into action_evals (id, parent_id, parent_type, eval_type, created_at, processed_at) VALUES (:id, :parent_id, :parent_type, :eval_type, :created_at, :processed_at)"
}

func (*postgresQueryEngine) actionInstances() string {
	return "select id, action_id, attempt, status, outcome, outcome_reason, created_at, updated_at, finished_at from action_instances where action_id=$1"
}

func (*postgresQueryEngine) getActionEval() string {
	return "select id, parent_id, parent_type, eval_type, created_at, processed_at from action_evals where id=$1"
}

func (*postgresQueryEngine) getActionInstance() string {
	return "select id, action_id, attempt, status, outcome, outcome_reason, created_at, updated_at, finished_at from action_instances where id=$1"
}

func (*postgresQueryEngine) actionInstancesByStatus() string {
	return "select id, action_id, attempt, status, outcome, outcome_reason, created_at, updated_at, finished_at from action_instances where status=$1"
}

func (*postgresQueryEngine) changeEventForObject() string {
	return "select mutation_id, mutation_time, event_type, object_id, object_type, parent_id, namespace, processed_at, experiment_payload, model_payload, model_version_payload, action_payload, action_instance_payload from mutation_events where object_id = $1"
}

type PostgresStorage struct {
	*SQLStorage
	*PostgresDriverUtils
	db     *sqlx.DB
	config *storageconfig.PostgresConfig

	logger *zap.Logger
}

func NewPostgresStorage(config *storageconfig.PostgresConfig, logger *zap.Logger) (*PostgresStorage, error) {
	db, err := sqlx.Open(POSTGRES_DRIVER, config.DataSource())
	if err != nil {
		return nil, fmt.Errorf("unable to connect to postgres %v", err)
	}
	sqlStorage := NewSQLStorage(db, &postgresQueryEngine{}, logger)
	util := &PostgresDriverUtils{db: db, config: config, logger: logger}
	return &PostgresStorage{SQLStorage: sqlStorage, PostgresDriverUtils: util, db: db, config: config, logger: logger}, nil
}

func (p *PostgresStorage) Ping() error {
	return p.db.Ping()
}

func (p *PostgresStorage) Close() error {
	return p.db.Close()
}

func (p *PostgresStorage) Backend() *BackendInfo {
	return &BackendInfo{Name: POSTGRES_DRIVER}
}
