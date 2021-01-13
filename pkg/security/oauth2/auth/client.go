package auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

/***********************************
	Abstraction
 ***********************************/
type OAuth2Client interface {
	ClientId() string
	SecretRequired() bool
	Secret() string
	GrantTypes() utils.StringSet
	RedirectUris() utils.StringSet
	Scoped() bool
	Scopes() utils.StringSet
	AutoApproveScopes() utils.StringSet
	AccessTokenValidity() time.Duration
	RereshTokenValidity() time.Duration
	UseSessionTimeout() bool
	TenantRestrictions() utils.StringSet
	//MaxTokensPerUser() int // TODO if this still needed?

	// TODO if resource id still needed?
}

/***********************************
	Default implementation
 ***********************************/
type ClientDetails struct {
	ClientId             string
	Secret               string
	GrantTypes           utils.StringSet
	RedirectUris         utils.StringSet
	Scopes               utils.StringSet
	AutoApproveScopes    utils.StringSet
	AccessTokenValidity  time.Duration
	RefreshTokenValidity time.Duration
	UseSessionTimeout    bool
	TenantRestrictions   utils.StringSet
}

// DefaultAouth2Client implements security.Account & OAuth2Client
type DefaultOAuth2Client struct {
	ClientDetails
}

// deja vu
func NewClient() *DefaultOAuth2Client {
	return &DefaultOAuth2Client{}
}

func NewClientWithDetails(clientDetails ClientDetails) *DefaultOAuth2Client {
	return &DefaultOAuth2Client{
		ClientDetails: clientDetails,
	}
}

/** OAuth2Client **/
func (c *DefaultOAuth2Client) ClientId() string {
	return c.ClientDetails.ClientId
}

func (c *DefaultOAuth2Client) SecretRequired() bool {
	return c.ClientDetails.Secret != ""
}

func (c *DefaultOAuth2Client) Secret() string {
	return c.ClientDetails.Secret
}

func (c *DefaultOAuth2Client) GrantTypes() utils.StringSet {
	return c.ClientDetails.GrantTypes
}

func (c *DefaultOAuth2Client) RedirectUris() utils.StringSet {
	return c.ClientDetails.RedirectUris
}

func (c *DefaultOAuth2Client) Scoped() bool {
	return c.ClientDetails.Scopes != nil && len(c.ClientDetails.Scopes) != 0
}

func (c *DefaultOAuth2Client) Scopes() utils.StringSet {
	return c.ClientDetails.Scopes
}

func (c *DefaultOAuth2Client) AutoApproveScopes() utils.StringSet {
	return c.ClientDetails.AutoApproveScopes
}

func (c *DefaultOAuth2Client) AccessTokenValidity() time.Duration {
	return c.ClientDetails.AccessTokenValidity
}

func (c *DefaultOAuth2Client) RereshTokenValidity() time.Duration {
	return c.ClientDetails.RefreshTokenValidity
}

func (c *DefaultOAuth2Client) UseSessionTimeout() bool {
	return c.ClientDetails.UseSessionTimeout
}

func (c *DefaultOAuth2Client) TenantRestrictions() utils.StringSet {
	return c.ClientDetails.TenantRestrictions
}

func (c *DefaultOAuth2Client) MaxTokensPerUser() int {
	return -1
}

/** security.Account **/
func (c *DefaultOAuth2Client) ID() interface{} {
	return c.ClientDetails.ClientId
}

func (c *DefaultOAuth2Client) Type() security.AccountType {
	return security.AccountTypeDefault
}

func (c *DefaultOAuth2Client) Username() string {
	return c.ClientDetails.ClientId
}

func (c *DefaultOAuth2Client) Credentials() interface{} {
	return c.ClientDetails.Secret
}

func (c *DefaultOAuth2Client) Permissions() []string {
	return c.ClientDetails.Scopes.Values()
}

func (c *DefaultOAuth2Client) Disabled() bool {
	return false
}

func (c *DefaultOAuth2Client) Locked() bool {
	return false
}

func (c *DefaultOAuth2Client) UseMFA() bool {
	return false
}