package service

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
)

type InMemoryApprovalStore struct {
	userApproval map[string][]*auth.Approval
}

func NewInMemoryApprovalStore() auth.ApprovalStore {
	return &InMemoryApprovalStore{
		userApproval: make(map[string][]*auth.Approval),
	}
}

func (m *InMemoryApprovalStore) SaveApproval(c context.Context, a *auth.Approval) error {
	approvals := m.userApproval[a.Username]
	approvals = append(approvals, a)
	m.userApproval[a.Username] = approvals
	return nil
}

func (m *InMemoryApprovalStore) LoadUserApprovalsByClientId(c context.Context, opts ...auth.ApprovalLoadOptions) ([]*auth.Approval, error) {
	opt := &auth.Approval{}
	for _, f := range opts {
		f(opt)
	}

	approvals := m.userApproval[opt.Username]
	var ret []*auth.Approval
	for _, a := range approvals {
		if a.ClientId == opt.ClientId {
			ret = append(ret, a)
		}
	}
	return ret, nil
}
