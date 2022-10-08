package membership

import (
	"crypto/sha1"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tensorland/modelbox/server/config"
	"github.com/tensorland/modelbox/server/utils"
	"go.uber.org/zap"
)

type ClusterMember struct {
	Id       string
	HostName string
	RPCAddr  string
	HTTPAddr string
}

func NewClusterMember(hostName, rpcAddr, httpAddr string) *ClusterMember {
	h := sha1.New()
	utils.HashString(h, hostName)
	utils.HashString(h, rpcAddr)
	utils.HashString(h, httpAddr)
	id := fmt.Sprintf("%x", h.Sum(nil))
	return &ClusterMember{
		Id:       id,
		HostName: hostName,
		RPCAddr:  rpcAddr,
		HTTPAddr: httpAddr,
	}
}

func (c ClusterMember) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c ClusterMember) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &c)
}

type ClusterMembership interface {
	// Gets the current list of members
	GetMembers() ([]*ClusterMember, error)

	// Asynchronously notify membership changes
	NotifyOnChange(cb func([]*ClusterMember))

	// Join the pool of servers
	Join() error

	// Leave the pool of servers
	Leave() error

	// Heartbeats once for testing
	renewOnce(t int64) error
}

func NewClusterMembership(svrConfig *config.ServerConfig, logger *zap.Logger) (ClusterMembership, error) {
	if svrConfig.ClusterMembershipBackend == "static" {
		members := []*ClusterMember{}
		for _, mem := range svrConfig.StaticClusterMembership.Members {
			members = append(members, &ClusterMember{Id: mem.Id, RPCAddr: mem.RPCAddr})
		}
		return &Static{members: members, logger: logger}, nil
	}

	if svrConfig.ClusterMembershipBackend == "mysql" {
	}
	return nil, fmt.Errorf("unable to create cluster membership driver for backend: %v", svrConfig.ClusterMembershipBackend)
}
