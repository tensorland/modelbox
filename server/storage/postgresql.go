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
