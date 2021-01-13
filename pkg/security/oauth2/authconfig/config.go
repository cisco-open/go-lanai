package authconfig

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
)

func ConfigureAuthorizationServer(registrar security.Registrar, configurer AuthorizationServerConfigurer) {
	config := &AuthorizationServerConfiguration{}
	configurer(config)

	registrar.Register(&TokenEndpointSecurityConfigurer{config: config})
}

type AuthorizationServerConfigurer func(*AuthorizationServerConfiguration)

type AuthorizationServerEndpoints struct {
	Authorize  string
	Token      string
	CheckToken string
	UserInfo   string
}

type AuthorizationServerConfiguration struct {
	ClientStore auth.OAuth2ClientStore
	ClientSecretEncoder passwd.PasswordEncoder
	Endpoints AuthorizationServerEndpoints
	errorHandler *auth.OAuth2ErrorHanlder
	// TODO
}

func (c *AuthorizationServerConfiguration) configureClientAuth(ws security.WebSecurity) {
	passwd.Configure(ws).
		AccountStore(c.clientAccountStore()).
		PasswordEncoder(c.clientSecretEncoder()).
		MFA(false)
	// no entry point, everything handled by access denied handler
	basicauth.Configure(ws).
		EntryPoint(nil)
	access.Configure(ws).
		Request(matcher.AnyRequest()).
		Authenticated()
	errorhandling.Configure(ws).
		AccessDeniedHandler(c.accessDeniedHandler()).
		AuthenticationErrorHandler(c.authenticationErrorHandler())
}

func (c *AuthorizationServerConfiguration) clientAccountStore() *auth.OAuth2ClientAccountStore{
	return auth.WrapOAuth2ClientStore(c.ClientStore)
}

func (c *AuthorizationServerConfiguration) clientSecretEncoder() passwd.PasswordEncoder {
	if c.ClientSecretEncoder == nil {
		c.ClientSecretEncoder = passwd.NewNoopPasswordEncoder()
	}
	return c.ClientSecretEncoder
}

func (c *AuthorizationServerConfiguration) accessDeniedHandler() security.AccessDeniedHandler {
	if c.errorHandler == nil {
		c.errorHandler = auth.NewOAuth2ErrorHanlder()
	}
	return c.errorHandler
}

func (c *AuthorizationServerConfiguration) authenticationErrorHandler() security.AuthenticationErrorHandler {
	if c.errorHandler == nil {
		c.errorHandler = auth.NewOAuth2ErrorHanlder()
	}
	return c.errorHandler
}


