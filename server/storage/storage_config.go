package storage

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

func (c *MySqlStorageConfig) dsnAdmin() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		c.UserName, c.Password, c.Host, strconv.Itoa(c.Port))
	return dsn
}
