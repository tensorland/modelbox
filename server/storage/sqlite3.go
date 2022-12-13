package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	storageconfig "github.com/tensorland/modelbox/server/storage/config"

	"go.uber.org/zap"
)

type sqliteQueryEngine struct{}

func (*sqliteQueryEngine) isDuplicate(err error) bool {
	if driverErr, ok := err.(*pq.Error); ok {
		fmt.Println("Diptanu error code ", driverErr.Code)
		if driverErr.Code == "23505" {
			return true
		}
	}
	return false
}

func (*sqliteQueryEngine) updateMetadata() string {
	return "INSERT INTO metadata (id, parent_id, metadata) VALUES(:id, :parent_id, :metadata) ON CONFLICT (id) DO UPDATE SET updated_at  = EXCLUDED.updated_at, metadata = EXCLUDED.metadata "
}

func (*sqliteQueryEngine) createExperiment() string {
	return "insert into experiments(id, name, owner, namespace, external_id, ml_framework, created_at, updated_at) values(:id, :name, :owner, :namespace, :external_id, :ml_framework, :created_at, :updated_at) ON CONFLICT (id) DO UPDATE SET updated_at = EXCLUDED.updated_at"
}

func (*sqliteQueryEngine) createCheckpoint() string {
	return "insert into checkpoints(id, experiment, epoch, metrics, created_at, updated_at) values(:id, :experiment, :epoch, :metrics, :created_at, :updated_at) ON CONFLICT (id) DO UPDATE SET updated_at = EXCLUDED.updated_at"
}

func (*sqliteQueryEngine) createModel() string {
	return "insert or ignore into models(id, name, owner, namespace, task, description, created_at, updated_at) values(:id, :name, :owner, :namespace, :task, :description, :created_at, :updated_at) ON CONFLICT (id) DO UPDATE SET updated_at = EXCLUDED.updated_at"
}

func (*sqliteQueryEngine) listEventsForObject() string {
	return "select id, parent_id, name, source_name, wallclock, metadata from events where parent_id=$1"
}

func (*sqliteQueryEngine) createModelVersion() string {
	return `insert or ignore into model_versions(id, name, model_id, version, description, ml_framework, unique_tags, created_at, updated_at) values(:id, :name, :model_id, :version, :description, :ml_framework, :unique_tags, :created_at, :updated_at)`
}

func (*sqliteQueryEngine) createAction() string {
	return "insert or ignore into actions (id, parent_id, name, arch, trigger_predicate, params, created_at, updated_at, finished_at) VALUES (:id, :parent_id, :name, :arch, :trigger_predicate, :params, :created_at, :updated_at, :finished_at)"
}

func (*sqliteQueryEngine) blobMultiWrite() string {
	return "insert or ignore into blobs(id, parent_id, metadata) VALUES "
}

func (*sqliteQueryEngine) actionInstances() string {
	return "select id, action_id, attempt, status, outcome, outcome_reason, created_at, updated_at, finished_at from action_instances where action_id=$1"
}

func (*sqliteQueryEngine) getActionInstance() string {
	return "select id, action_id, attempt, status, outcome, outcome_reason, created_at, updated_at, finished_at from action_instances where id=$1"
}

func (*sqliteQueryEngine) actionInstancesByStatus() string {
	return "select id, action_id, attempt, status, outcome, outcome_reason, created_at, updated_at, finished_at from action_instances where status=$1"
}

func (*sqliteQueryEngine) changeEventForObject() string {
	return "select mutation_id, mutation_time, event_type, object_id, object_type, parent_id, namespace, processed_at, experiment_payload, model_payload, model_version_payload, action_payload, action_instance_payload from mutation_events where object_id = $1"
}

type Sqlite3Storage struct {
	*SQLStorage
	config *storageconfig.Sqlite3Config
	db     *sqlx.DB
	logger *zap.Logger
}

func NewSqlite3Storage(config *storageconfig.Sqlite3Config, logger *zap.Logger) (*Sqlite3Storage, error) {
	db, err := sqlx.Open("sqlite3", config.DataSource())
	if err != nil {
		return nil, fmt.Errorf("unable to connect to sqlite3: %v", err)
	}
	sqlStorage := NewSQLStorage(db, &sqliteQueryEngine{}, logger)
	return &Sqlite3Storage{SQLStorage: sqlStorage, db: db, config: config, logger: logger}, nil
}

func (*SQLStorage) Backend() *BackendInfo {
	return &BackendInfo{
		Name: "sqlite3",
	}
}

func (*Sqlite3Storage) Close() error {
	return nil
}

func (s *Sqlite3Storage) CreateSchema(path string) error {
	files, err := filepath.Glob(fmt.Sprintf("%s/schema_ver*", path))
	if err != nil {
		return fmt.Errorf("unable to create schema: %v", err)
	}
	for _, file := range files {
		buf, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("unable to read schema file: %v", err)
		}
		queries := strings.Split(string(buf), ";")
		for _, query := range queries {
			if strings.TrimSpace(query) == "" {
				continue
			}
			if _, err := s.db.Exec(query); err != nil {
				return fmt.Errorf("unable to execute query: %v, err: %v", query, err)
			}
		}
	}
	return nil
}

func (*Sqlite3Storage) Ping() error {
	return nil
}
