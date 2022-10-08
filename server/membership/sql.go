package membership

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tensorland/modelbox/server/storage"
	storageconfig "github.com/tensorland/modelbox/server/storage/config"
	"go.uber.org/zap"
)

type clusterMembersSchema struct {
	Id            string
	Info          *ClusterMember
	HeartBeatTime int64 `db:"heartbeat_time"`
}

type queryRegistry interface {
	renewHeartbeat() string
	getMembers(staleDuration time.Duration) string
}

type mysqlQueryRegistry struct{}

func (*mysqlQueryRegistry) renewHeartbeat() string {
	return "INSERT INTO cluster_members(id, info, heartbeat_time) VALUES(:id, :info, :heartbeat_time) ON DUPLICATE KEY UPDATE heartbeat_time=:heartbeat_time"
}

func (*mysqlQueryRegistry) getMembers(staleDuration time.Duration) string {
	return ""
}

type SQLConfig struct {
	HBFrequency               time.Duration
	MaxStaleHeartBeatDuration time.Duration
}

type SQLMembership struct {
	config        *SQLConfig
	member        *ClusterMember
	queryRegistry queryRegistry
	db            *sqlx.DB
	stopCh        chan struct{}
	logger        *zap.Logger
}

func NewSQLMembership(config *SQLConfig, db *sqlx.DB, queryRegistry queryRegistry, member *ClusterMember, logger *zap.Logger) *SQLMembership {
	return &SQLMembership{config: config, db: db, queryRegistry: queryRegistry, member: member, stopCh: make(chan struct{}), logger: logger}
}

// Gets the current list of members
func (s *SQLMembership) GetMembers() ([]*ClusterMember, error) {
	members := []*ClusterMember{}
	ctx := context.Background()
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		rows := []clusterMembersSchema{}
		if err := tx.SelectContext(ctx, &rows, s.queryRegistry.getMembers(s.config.MaxStaleHeartBeatDuration)); err != nil {
			return err
		}
		for _, row := range rows {
			members = append(members, &ClusterMember{
				Id:       row.Id,
				HostName: row.Info.HostName,
				RPCAddr:  row.Info.RPCAddr,
				HTTPAddr: row.Info.HTTPAddr,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return members, nil
}

// Asynchronously notify membership changes
func (*SQLMembership) NotifyOnChange(cb func([]*ClusterMember)) {
}

func (s *SQLMembership) transact(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// Join the pool of servers
func (s *SQLMembership) Join() error {
	s.logger.Sugar().Infof("starting cluster membership. heartbeat frequency: %v", s.config.HBFrequency)
	s.heartBeat()
	return nil
}

// Leave the pool of servers
func (s *SQLMembership) Leave() error {
	close(s.stopCh)
	return nil
}

func (s *SQLMembership) heartBeat() {
	next := time.After(s.config.HBFrequency)
	for {
		select {
		case <-s.stopCh:
			s.logger.Sugar().Info("stopping to renew leases")
			return
		case <-next:
			s.renewOnce(time.Now().Unix())
		}
	}
}

func (s *SQLMembership) renewOnce(t int64) error {
	ctx := context.Background()
	schema := &clusterMembersSchema{
		Id:            s.member.Id,
		Info:          s.member,
		HeartBeatTime: t,
	}
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.NamedExecContext(ctx, s.queryRegistry.renewHeartbeat(), schema)
		return err
	})
	if err != nil {
		s.logger.Sugar().Errorf("unable to renew heartbeat: %v", err)
	}
	return err
}

type MysqlClusterMembership struct {
	*SQLMembership
	*storage.MYSQLDriverUtils
}

func NewMysqlClusterMembership(sqlConfig *SQLConfig, member *ClusterMember, config *storageconfig.MySqlStorageConfig, logger *zap.Logger) (*MysqlClusterMembership, error) {
	db, err := sqlx.Open("mysql", config.DataSource())
	if err != nil {
		return nil, fmt.Errorf("unable to connect to mysql %v", err)
	}
	membership := NewSQLMembership(sqlConfig, db, &mysqlQueryRegistry{}, member, logger)
	util := &storage.MYSQLDriverUtils{Config: config, Db: db, Logger: logger}
	return &MysqlClusterMembership{SQLMembership: membership, MYSQLDriverUtils: util}, nil
}

type PostgresClusterMembership struct {
	*SQLMembership
}
