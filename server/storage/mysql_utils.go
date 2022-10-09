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

type MYSQLDriverUtils struct {
	Config *storageconfig.MySqlStorageConfig
	Db     *sqlx.DB
	Logger *zap.Logger
}

func (m *MYSQLDriverUtils) CreateSchema(path string) error {
	db, err := sqlx.Open("mysql", m.Config.DsnAdmin())
	if err != nil {
		return fmt.Errorf("unable to connec to db: %v", err)
	}
	defer db.Close()
	queries := []string{
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", m.Config.DbName),
		fmt.Sprintf("USE %s", m.Config.DbName),
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
	m.Logger.Sugar().Infof("applying the following schema files: %v", strings.Join(files, ","))
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
			if _, err := m.Db.Exec(query); err != nil {
				return fmt.Errorf("unable to execute query: %v", err)
			}
		}
	}
	return nil
}

func (m *MYSQLDriverUtils) DropDb() error {
	db, err := sqlx.Open("mysql", m.Config.DsnAdmin())
	if err != nil {
		return fmt.Errorf("unable to connec to db: %v", err)
	}
	defer db.Close()
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", m.Config.DbName))
	return err
}
