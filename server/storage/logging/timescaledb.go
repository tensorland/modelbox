package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	LOG_INSERT = "INSERT INTO metrics(time, parent_id, name, double_value, tensor, step, wallclock) VALUES(NOW(), :parent_id, :name, :double_value, :tensor, :step, :wallclock)"

	LOG_GET = "SELECT time, parent_id, name, double_value, tensor, step, wallclock FROM metrics where parent_id=$1"
)

type MetricsSchema struct {
	Time      time.Time `db:"time"`
	ParentId  string    `db:"parent_id"`
	Name      string    `db:"name"`
	FloatVal  float32   `db:"double_value"`
	TensorVal string    `db:"tensor"`
	Step      uint64    `db:"step"`
	Wallclock uint64    `db:"wallclock"`
}

type TimescaleDbConfig struct {
	Host     string
	Port     int
	UserName string
	Password string
	DbName   string
}

func (p *TimescaleDbConfig) DataSource() string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.UserName, p.Password, p.DbName)
	return dsn
}

func (p *TimescaleDbConfig) dsnAdmin() string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.UserName, p.Password, "postgres")
	return dsn
}

type TimescaleDbLogger struct {
	db *sqlx.DB

	config *TimescaleDbConfig
	logger *zap.Logger
}

func NewTimescaleDbLogger(config *TimescaleDbConfig, logger *zap.Logger) (*TimescaleDbLogger, error) {
	db, err := sqlx.Open("postgres", config.DataSource())
	if err != nil {
		return nil, fmt.Errorf("unable to connect to timescaledb %v", err)
	}
	return &TimescaleDbLogger{
		db:     db,
		config: config,
		logger: logger,
	}, nil
}

func (t *TimescaleDbLogger) LogFloats(ctx context.Context, parentId, key string, value *FloatLog) error {
	row := &MetricsSchema{
		Time:      time.Now(),
		ParentId:  parentId,
		Name:      key,
		FloatVal:  value.Value,
		Step:      value.Step,
		Wallclock: value.WallClock,
	}
	return t.transact(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.NamedExecContext(ctx, LOG_INSERT, row)
		return err
	})
}

func (t *TimescaleDbLogger) GetFloatLogs(ctx context.Context, parentId string) (map[string][]*FloatLog, error) {
	logs := map[string][]*FloatLog{}
	err := t.transact(ctx, func(tx *sqlx.Tx) error {
		foo := []MetricsSchema{}
		if err := tx.SelectContext(ctx, &foo, LOG_GET, parentId); err != nil {
			return fmt.Errorf("can't retrieve metrics for object: %v, err: %v", parentId, err)
		}
		for _, row := range foo {
			log := &FloatLog{Value: row.FloatVal, Step: row.Step, WallClock: row.Wallclock}
			logs[row.Name] = append(logs[row.Name], log)
		}
		return nil
	})
	return logs, err
}

func (t *TimescaleDbLogger) transact(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (t *TimescaleDbLogger) Close() error {
	return t.db.Close()
}

func (t *TimescaleDbLogger) DropDb() error {
	db, err := sqlx.Open("postgres", t.config.dsnAdmin())
	if err != nil {
		return fmt.Errorf("unable to connec to db: %v", err)
	}
	defer db.Close()

	t.dropConnections(t.db, t.config.DbName)
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", t.config.DbName))
	return err
}

func (t *TimescaleDbLogger) CreateSchema(path string) error {
	db, dbCloser, err := t.connect(t.config.dsnAdmin())
	if err != nil {
		return err
	}
	defer dbCloser()
	t.dropConnections(db, t.config.DbName)
	queries := []string{
		fmt.Sprintf("DROP DATABASE IF EXISTS %v", t.config.DbName),
		fmt.Sprintf("CREATE DATABASE %v", t.config.DbName),
	}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("unable to execute query: %v err: %v", query, err)
		}
	}
	files, err := filepath.Glob(fmt.Sprintf("%s/schema_ver*", path))
	if err != nil {
		return fmt.Errorf("unable to read schema files: %v", err)
	}
	sort.Strings(files)
	t.logger.Sugar().Info("applying the following schema files: ", strings.Join(files, ","))
	for _, file := range files {
		buf, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("unable to read schema: %v", err)
		}
		queries = strings.Split(string(buf), ";")
		for _, query := range queries {
			if strings.TrimSpace(query) == "" {
				continue
			}
			if _, err := t.db.Exec(query); err != nil {
				return fmt.Errorf("unable to execute query: %v, err: %v", query, err)
			}
		}
	}
	return nil
}

func (t *TimescaleDbLogger) connect(source string) (*sqlx.DB, func(), error) {
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

func (t *TimescaleDbLogger) dropConnections(db *sqlx.DB, name string) {
	query := `
		select pg_terminate_backend(pg_stat_activity.pid)
		from pg_stat_activity
		where pg_stat_activity.datname = $1 and pid <> pg_backend_pid()`
	_, err := db.Exec(query, name)
	if err != nil {
		t.logger.Sugar().Errorf("error dropping connections: %v", err)
	}
}

func (*TimescaleDbLogger) Backend() string {
	return "timescaledb"
}
