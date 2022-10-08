package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	storageconfig "github.com/tensorland/modelbox/server/storage/config"
	"go.uber.org/zap"
)

type PostgresDriverUtils struct {
	db     *sqlx.DB
	config *storageconfig.PostgresConfig

	logger *zap.Logger
}

func (p *PostgresDriverUtils) DropDb() error {
	db, err := sqlx.Open(POSTGRES_DRIVER, p.config.DsnAdmin())
	if err != nil {
		return fmt.Errorf("unable to connec to db: %v", err)
	}
	defer db.Close()

	p.dropConnections(p.db, p.config.DbName)
	p.db.Close()
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", p.config.DbName))
	return err
}

func (p *PostgresDriverUtils) CreateSchema(path string) error {
	db, dbCloser, err := p.connect(p.config.DsnAdmin())
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
	files, err := filepath.Glob(fmt.Sprintf("%s/schema_ver*", path))
	if err != nil {
		return fmt.Errorf("unable to read schema files: %v", err)
	}
	sort.Strings(files)
	p.logger.Sugar().Info("applying the following schema files: ", strings.Join(files, ","))
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
			if _, err := p.db.Exec(query); err != nil {
				return fmt.Errorf("unable to execute query: %v, err: %v", query, err)
			}
		}
	}
	return nil
}

func (p *PostgresDriverUtils) connect(source string) (*sqlx.DB, func(), error) {
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

func (p *PostgresDriverUtils) dropConnections(db *sqlx.DB, name string) {
	query := `
		select pg_terminate_backend(pg_stat_activity.pid)
		from pg_stat_activity
		where pg_stat_activity.datname = $1 and pid <> pg_backend_pid()`
	_, err := db.Exec(query, name)
	if err != nil {
		p.logger.Sugar().Errorf("error dropping connections: %v", err)
	}
}
