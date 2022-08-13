package storage

import (
	"fmt"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	POSTGRES_DRIVER = "postgres"
)

type postgresDriverUtils struct{}

func (*postgresDriverUtils) isDuplicate(err error) bool {
	if driverErr, ok := err.(*pq.Error); ok {
		fmt.Println("Diptanu ", driverErr.Code)
		if driverErr.Code == "23505" {
			return true
		}
	}
	return false
}

func (*postgresDriverUtils) updateMetadata() string {
	return "INSERT INTO metadata (id, parent_id, metadata) VALUES(:id, :parent_id, :metadata) ON CONFLICT (id) DO UPDATE SET updated_at  = EXCLUDED.updated_at, metadata = EXCLUDED.metadata "
}

func (*postgresDriverUtils) createExperiment() string {
	return "insert into experiments(id, name, owner, namespace, external_id, ml_framework, created_at, updated_at) values(:id, :name, :owner, :namespace, :external_id, :ml_framework, :created_at, :updated_at) ON CONFLICT (id) DO UPDATE SET updated_at = EXCLUDED.updated_at"
}

func (*postgresDriverUtils) createCheckpoint() string {
	return "insert into checkpoints(id, experiment, epoch, metrics, created_at, updated_at) values(:id, :experiment, :epoch, :metrics, :created_at, :updated_at) ON CONFLICT (id) DO UPDATE SET updated_at = EXCLUDED.updated_at"
}

func (*postgresDriverUtils) createModel() string {
	return "insert into models(id, name, owner, namespace, task, description, created_at, updated_at) values(:id, :name, :owner, :namespace, :task, :description, :created_at, :updated_at) ON CONFLICT (id) DO UPDATE SET updated_at = EXCLUDED.updated_at"
}

type PostgresStorage struct {
	*SQLStorage
	db     *sqlx.DB
	config *PostgresConfig

	logger *zap.Logger
}

func NewPostgresStorage(config *PostgresConfig, logger *zap.Logger) (*PostgresStorage, error) {
	db, err := sqlx.Open(POSTGRES_DRIVER, config.DataSource())
	if err != nil {
		return nil, fmt.Errorf("unable to connect to postgres %v", err)
	}
	sqlStorage := NewSQLStorage(db, &postgresDriverUtils{}, logger)
	return &PostgresStorage{SQLStorage: sqlStorage, db: db, config: config, logger: logger}, nil
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

func (p *PostgresStorage) DropDb() error {
	db, err := sqlx.Open(POSTGRES_DRIVER, p.config.dsnAdmin())
	if err != nil {
		return fmt.Errorf("unable to connec to db: %v", err)
	}
	defer db.Close()

	p.dropConnections(p.db, p.config.DbName)
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", p.config.DbName))
	return err
}

func (p *PostgresStorage) CreateSchema(path string) error {
	db, dbCloser, err := p.connect(p.config.dsnAdmin())
	if err != nil {
		return err
	}
	defer dbCloser()
	p.dropConnections(db, p.config.DbName)
	queries := []string{
		fmt.Sprintf("DROP DATABASE IF EXISTS %v", p.config.DbName),
		fmt.Sprintf("CREATE DATABASE %v", p.config.DbName),
	}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("unable to execute query: %v err: %v", query, err)
		}
	}
	buf, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("unable to read schema: %v", err)
	}
	queries = strings.Split(string(buf), ";")
	for _, query := range queries {
		if strings.TrimSpace(query) == "" {
			continue
		}
		if _, err := p.db.Exec(query); err != nil {
			return fmt.Errorf("unable to execute query: %v, err: %v", query, err)
		}
	}
	return nil
}

func (p *PostgresStorage) connect(source string) (*sqlx.DB, func(), error) {
	db, err := sqlx.Open("postgres", source)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to open db: %v", err)
	}
	return db, func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}, nil
}

func (p *PostgresStorage) dropConnections(db *sqlx.DB, name string) {
	query := `
		select pg_terminate_backend(pg_stat_activity.pid)
		from pg_stat_activity
		where pg_stat_activity.datname = $1 and pid <> pg_backend_pid()`
	_, err := db.Exec(query, name)
	if err != nil {
		p.logger.Sugar().Errorf("error dropping connections: %v", err)
	}
}
