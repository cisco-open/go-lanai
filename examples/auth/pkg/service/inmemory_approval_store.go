package service

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
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

func (m *InMemoryApprovalStore) SaveApproval(c context.Context, user security.Account, a *auth.Approval) error {
	approvals := m.userApproval[user.Username()]
	approvals = append(approvals, a)
	m.userApproval[user.Username()] = approvals
	return nil
}

func (m *InMemoryApprovalStore) LoadApprovalsByClientId(c context.Context, user security.Account, clientId string) ([]*auth.Approval, error) {
	approvals := m.userApproval[user.Username()]
	var ret []*auth.Approval
	for _, a := range approvals {
		if a.ClientId == clientId {
			ret = append(ret, a)
		}
	}
	return ret, nil
}
