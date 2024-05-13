package auth

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/utils"
)

type Approval struct {
	UserId      interface{}
	Username    string
	ClientId    string
	RedirectUri string
	Scopes      utils.StringSet
}

type ApprovalLoadOptions func(*Approval)

type ApprovalStore interface {
	SaveApproval(c context.Context, a *Approval) error
	LoadUserApprovalsByClientId(c context.Context, opts ...ApprovalLoadOptions) ([]*Approval, error)
}

func WithUserId(userId interface{}) ApprovalLoadOptions {
	return func(a *Approval) {
		a.UserId = userId
	}
}

func WithUsername(username string) ApprovalLoadOptions {
	return func(a *Approval) {
		a.Username = username
	}
}

func WithClientId(clientId string) ApprovalLoadOptions {
	return func(a *Approval) {
		a.ClientId = clientId
	}
}
