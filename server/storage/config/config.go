package config

import (
	"fmt"
	"strconv"
)

type MySqlStorageConfig struct {
	Host     string
	Port     int
	UserName string
	Password string
	DbName   string
}

func (c *MySqlStorageConfig) DataSource() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8",
		c.UserName, c.Password, c.Host, strconv.Itoa(c.Port), c.DbName)
	return dsn
}

func (c *MySqlStorageConfig) DsnAdmin() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		c.UserName, c.Password, c.Host, strconv.Itoa(c.Port))
	return dsn
}

type PostgresConfig struct {
	Host     string
	Port     int
	UserName string
	Password string
	DbName   string
}

func (p *PostgresConfig) DataSource() string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.UserName, p.Password, p.DbName)
	return dsn
}

func (p *PostgresConfig) DsnAdmin() string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.UserName, p.Password, "postgres")
	return dsn
}

type Sqlite3Config struct {
	File     string
	InMemory bool
}

func (s *Sqlite3Config) DataSource() string {
	dsn := fmt.Sprintf("%v", s.File)
	return dsn
}

func (s *Sqlite3Config) DsnAdmin() string {
	return s.DataSource()
}
