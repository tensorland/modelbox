package storage

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

var (
	user   string
	pass   string
	prot   string
	host   string
	port   uint64
	dbname string
)

func init() {
	// get environment variables
	env := func(key, defaultValue string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return defaultValue
	}
	port, _ = strconv.ParseUint(env("MYSQL_TEST_PORT", "3306"), 10, 64)
	user = env("MYSQL_TEST_USER", "root")
	pass = env("MYSQL_TEST_PASS", "foo")
	prot = env("MYSQL_TEST_PROT", "tcp")
	host = env("MYSQL_TEST_HOST", "172.17.0.2")
	dbname = env("MYSQL_TEST_DBNAME", "gotest")
}

type MySQLTestSuite struct {
	suite.Suite
	StorageInterfaceTestSuite
	mysql *MySqlStorage
}

func (s *MySQLTestSuite) SetupSuite() {
	if err := s.mysql.CreateSchema("schemas/mysql/schema_ver_0.sql"); err != nil {
		s.T().Fatalf("unable to load mysql schema: %v", err)
	}
}

func (s *MySQLTestSuite) TearDownSuite() {
	if err := s.mysql.DropDb(); err != nil {
		s.T().Fatalf("unable to drop db: %v", err)
	}
}

func TestMySqlTestSuite(t *testing.T) {
	logger, _ := zap.NewProduction()
	config := MySqlStorageConfig{
		Host:     host,
		Port:     int(port),
		UserName: user,
		Password: pass,
		DbName:   dbname,
	}
	mysql, err := NewMySqlStorage(&config, logger)
	if err != nil {
		t.Fatalf("unable to create mysql: %v", err)
	}
	suite.Run(t, &MySQLTestSuite{
		StorageInterfaceTestSuite: StorageInterfaceTestSuite{
			t:         t,
			storageIf: mysql,
		},
		mysql: mysql,
	})
}
