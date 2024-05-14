package testdata

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
)

type MockedApprovalStore struct {
	userApproval map[string][]*auth.Approval
}

func NewMockedApprovalStore() auth.ApprovalStore {
	return &MockedApprovalStore{
		userApproval: make(map[string][]*auth.Approval),
	}
}

func (m *MockedApprovalStore) SaveApproval(c context.Context, a *auth.Approval) error {
	approvals := m.userApproval[a.Username]
	approvals = append(approvals, a)
	m.userApproval[a.Username] = approvals
	return nil
}

func (m *MockedApprovalStore) LoadApprovals(c context.Context, opts ...auth.ApprovalLoadOptions) ([]*auth.Approval, error) {
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
