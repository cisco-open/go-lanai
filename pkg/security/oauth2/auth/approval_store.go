package auth

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/utils"
)

type Approval struct {
	ClientId    string
	RedirectUri string
	Scopes      utils.StringSet
}

type ApprovalStore interface {
	SaveApproval(c context.Context, user security.Account, a *Approval) error
	LoadApprovalsByClientId(c context.Context, user security.Account, clientId string) ([]*Approval, error)
}
