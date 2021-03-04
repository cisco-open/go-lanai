package oauth2

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

/***********************************
	DTO
 ***********************************/
type OAuth2Client interface {
	ClientId() string
	SecretRequired() bool
	Secret() string
	GrantTypes() utils.StringSet
	RedirectUris() utils.StringSet
	Scopes() utils.StringSet
	AutoApproveScopes() utils.StringSet
	AccessTokenValidity() time.Duration
	RereshTokenValidity() time.Duration
	UseSessionTimeout() bool
	TenantRestrictions() utils.StringSet
	ResourceIDs() utils.StringSet
	//MaxTokensPerUser() int // TODO if this still needed?
}

/***********************************
	Store
 ***********************************/
type OAuth2ClientStore interface {
	LoadClientByClientId(ctx context.Context, clientId string) (OAuth2Client, error)
}

