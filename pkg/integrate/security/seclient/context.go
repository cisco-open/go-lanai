package seclient

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

type AuthOptions func(opt *AuthOption)

type AuthOption struct {
	Password    string	// Password is used by password login
	AccessToken string	// AccessToken is used by switch user/tenant
	Username    string	// Username is used by password login and switch user
	UserId      string	// UserId is used by switch user
	TenantId    string	// TenantId is used by password login and switch user/tenant
	TenantName  string	// TenantName is used by password login and switch user/tenant

}

type AuthenticationClient interface {
	PasswordLogin(ctx context.Context, opts ...AuthOptions) (*Result, error)
	SwitchUser(ctx context.Context, opts ...AuthOptions) (*Result, error)
	SwitchTenant(ctx context.Context, opts ...AuthOptions) (*Result, error)
}

type Result struct {
	Request oauth2.OAuth2Request
	Token oauth2.AccessToken
}

/****************************
	AuthOptions
 ****************************/

func WithCredentials(username, password string) AuthOptions {
	return func(opt *AuthOption) {
		opt.Username = username
		opt.Password = password
	}
}

func WithCurrentSecurity(ctx context.Context) AuthOptions {
	return WithAuthentication(security.Get(ctx))
}

func WithAuthentication(auth security.Authentication) AuthOptions {
	oauth, ok := auth.(oauth2.Authentication)
	if !ok {
		return func(opt *AuthOption) {}
	}
	return WithAccessToken(oauth.AccessToken().Value())
}

func WithAccessToken(accessToken string) AuthOptions {
	return func(opt *AuthOption) {
		opt.AccessToken = accessToken
	}
}

// WithTenant create an options that specify tenant by either tenantId or tenantName
// username and userId are exclusive, cannot be both empty
func WithTenant(tenantId string, tenantName string) AuthOptions {
	if tenantId != "" {
		return WithTenantId(tenantId)
	} else {
		return WithTenantName(tenantName)
	}
}

func WithTenantId(tenantId string) AuthOptions {
	return func(opt *AuthOption) {
		opt.TenantId = tenantId
		opt.TenantName = ""
	}
}

func WithTenantName(tenantName string) AuthOptions {
	return func(opt *AuthOption) {
		opt.TenantId = ""
		opt.TenantName = tenantName
	}
}

// WithUser create an options that specify user by either username or userId
// username and userId are exclusive, cannot be both empty
func WithUser(userId string, username string) AuthOptions {
	if username != "" {
		return WithUsername(username)
	} else {
		return WithUserId(userId)
	}
}

func WithUsername(username string) AuthOptions {
	return func(opt *AuthOption) {
		opt.Username = username
		opt.UserId = ""
	}
}

func WithUserId(userId string) AuthOptions {
	return func(opt *AuthOption) {
		opt.Username = ""
		opt.UserId = userId
	}
}