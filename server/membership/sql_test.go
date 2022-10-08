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

func (s *ClusterMembershipTestSuite) TestJoin(t *testing.T) {
	currentTime := time.Now().Unix()
	err := s.membership.renewOnce(currentTime)
	assert.Nil(s.t, err)
}
