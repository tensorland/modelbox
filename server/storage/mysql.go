package storage

import (
	"fmt"
	"os"
	"strings"

	"github.com/VividCortex/mysqlerr"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

const (
	MYSQL_DRIVER = "mysql"
)

type mySQLDriverUtils struct{}

func (*mySQLDriverUtils) isDuplicate(err error) bool {
	if driverErr, ok := err.(*mysql.MySQLError); ok {
		if driverErr.Number == mysqlerr.ER_DUP_ENTRY {
			return true
		}
	}
	return false
}

func (*mySQLDriverUtils) updateMetadata() string {
	return "INSERT INTO metadata (id, parent_id, metadata) VALUES(:id, :parent_id, :metadata) ON DUPLICATE KEY UPDATE id=VALUES(`id`), parent_id=VALUES(`parent_id`), metadata=VALUES(`metadata`)"
}

func (*mySQLDriverUtils) createExperiment() string {
	return "insert into experiments(id, name, owner, namespace, external_id, ml_framework, created_at, updated_at) values(:id, :name, :owner, :namespace, :external_id, :ml_framework, :created_at, :updated_at)"
}

func (*mySQLDriverUtils) createCheckpoint() string {
	return "insert into checkpoints(id, experiment, epoch, metrics, created_at, updated_at) values(:id, :experiment, :epoch, :metrics, :created_at, :updated_at)"
}

func (*mySQLDriverUtils) createModel() string {
	return "insert into models(id, name, owner, namespace, task, description, created_at, updated_at) values(:id, :name, :owner, :namespace, :task, :description, :created_at, :updated_at)"
}

type MySqlStorage struct {
	*SQLStorage
	db     *sqlx.DB
	config *MySqlStorageConfig

	logger *zap.Logger
}

func NewMySqlStorage(config *MySqlStorageConfig, logger *zap.Logger) (*MySqlStorage, error) {
	db, err := sqlx.Open(MYSQL_DRIVER, config.DataSource())
	if err != nil {
		return nil, fmt.Errorf("unable to connect to mysql %v", err)
	}
	sqlStorage := NewSQLStorage(db, &mySQLDriverUtils{}, logger)
	return &MySqlStorage{SQLStorage: sqlStorage, db: db, config: config, logger: logger}, nil
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

func (e *MySqlStorage) DropDb() error {
	db, err := sqlx.Open(MYSQL_DRIVER, e.config.dsnAdmin())
	if err != nil {
		return fmt.Errorf("unable to connec to db: %v", err)
	}
	defer db.Close()
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", e.config.DbName))
	return err
}

func (e *MySqlStorage) CreateSchema(path string) error {
	db, err := sqlx.Open(MYSQL_DRIVER, e.config.dsnAdmin())
	if err != nil {
		return fmt.Errorf("unable to connec to db: %v", err)
	}
	defer db.Close()
	queries := []string{
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", e.config.DbName),
		fmt.Sprintf("USE %s", e.config.DbName),
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
		if _, err := e.db.Exec(query); err != nil {
			return fmt.Errorf("unable to execute query: %v", err)
		}
	}
	return nil
}
