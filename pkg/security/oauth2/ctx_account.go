package oauth2

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
)

//goland:noinspection GoNameStartsWithPackageName
type OAuth2AccountStore interface {
	// LoadAccountById find account by its Domain
	LoadAccountById(ctx context.Context, id interface{}, clientId string) (security.Account, error)
	// LoadAccountByUsername find account by its Username
	LoadAccountByUsername(ctx context.Context, username string, clientId string) (security.Account, error)
}
