package membership

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type ClusterMembershipTestSuite struct {
	t *testing.T

	membership ClusterMembership
}

func (s *ClusterMembershipTestSuite) TestJoin() {
	currentTime := time.Now().Unix()
	err := s.membership.renewOnce(currentTime)
	assert.Nil(s.t, err)

	// Get the member
	members, err := s.membership.GetMembers()
	assert.Nil(s.t, err)
	assert.Equal(s.t, 1, len(members))

	assert.Equal(s.t, "host1", members[0].HostName)
	assert.Equal(s.t, "192.168.1.2:8085", members[0].RPCAddr)
	assert.Equal(s.t, "192.168.1.2:8090", members[0].HTTPAddr)
}
