package membership

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	storageconfig "github.com/tensorland/modelbox/server/storage/config"
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
	host = env("MYSQL_TEST_HOST", "172.20.0.6")
	dbname = env("MYSQL_TEST_DBNAME", "gotest")
}

type MySQLMembershipTestSuite struct {
	suite.Suite
	*ClusterMembershipTestSuite
	mysql *MysqlClusterMembership
}

func (s *MySQLMembershipTestSuite) SetupSuite() {
	if err := s.mysql.CreateSchema("../storage/schemas/mysql/"); err != nil {
		s.T().Fatalf("unable to load mysql schema: %v", err)
	}
}

func (s *MySQLMembershipTestSuite) TearDownSuite() {
	if err := s.mysql.DropDb(); err != nil {
		s.T().Fatalf("unable to drop db: %v", err)
	}
}

func TestMySqlClusterMembershipTestSuite(t *testing.T) {
	logger, _ := zap.NewProduction()
	config := storageconfig.MySqlStorageConfig{
		Host:     host,
		Port:     int(port),
		UserName: user,
		Password: pass,
		DbName:   dbname,
	}
	hbConfig := &SQLConfig{
		HBFrequency:               5 * time.Second,
		MaxStaleHeartBeatDuration: 60 * time.Second,
	}
	member := NewClusterMember("host1", "192.168.1.2:8085", "192.168.1.2:8090")
	mysql, err := NewMysqlClusterMembership(hbConfig, member, &config, logger)
	if err != nil {
		t.Fatalf("unable to create mysql: %v", err)
	}
	suite.Run(t, &MySQLMembershipTestSuite{
		ClusterMembershipTestSuite: &ClusterMembershipTestSuite{
			t:          t,
			membership: mysql,
		},
		mysql: mysql,
	})
}
