package oauth2

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

/***********************************
	DTO
 ***********************************/

//goland:noinspection GoNameStartsWithPackageName
type OAuth2Client interface {
	ClientId() string
	SecretRequired() bool
	Secret() string
	GrantTypes() utils.StringSet
	RedirectUris() utils.StringSet
	Scopes() utils.StringSet
	AutoApproveScopes() utils.StringSet
	AccessTokenValidity() time.Duration
	RefreshTokenValidity() time.Duration
	UseSessionTimeout() bool
	TenantRestrictions() utils.StringSet
	ResourceIDs() utils.StringSet
}

/***********************************
	Store
 ***********************************/

//goland:noinspection GoNameStartsWithPackageName
type OAuth2ClientStore interface {
	LoadClientByClientId(ctx context.Context, clientId string) (OAuth2Client, error)
}

