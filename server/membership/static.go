package membership

import "go.uber.org/zap"

type Static struct {
	members []*ClusterMember
	logger  *zap.Logger
}

func NewStatic(logger *zap.Logger, members []*ClusterMember) *Static {
	return &Static{members: members, logger: logger}
}

// Gets the current list of members
func (s *Static) GetMembers() ([]*ClusterMember, error) {
	return s.members, nil
}

// Asynchronously notify membership changes
func (*Static) NotifyOnChange(cb func([]*ClusterMember)) {
	// Static list never changes
}

// Join the pool of servers
func (*Static) Join() error {
	// Members are already assumed to be joined
	return nil
}

// Leave the pool of servers
func (*Static) Leave() error {
	// Members never leave
	return nil
}

// Since we have a static list at the start this has no effect
func (*Static) renewOnce(t int64) error {
	return nil
}
