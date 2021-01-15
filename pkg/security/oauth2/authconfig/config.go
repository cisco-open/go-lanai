package authconfig

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/grants"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
)

type AuthorizationServerConfigurer func(*AuthorizationServerConfiguration)

// Configuration entry point
func ConfigureAuthorizationServer(registrar security.Registrar, configurer AuthorizationServerConfigurer) {
	config := &AuthorizationServerConfiguration{}
	configurer(config)

	registrar.Register(&ClientAuthEndpointsConfigurer{config: config})
}

/****************************
	configuration
 ****************************/
type AuthorizationServerEndpoints struct {
	Authorize  string
	Token      string
	CheckToken string
	UserInfo   string
}

type AuthorizationServerConfiguration struct {
	ClientStore         auth.OAuth2ClientStore
	ClientSecretEncoder passwd.PasswordEncoder
	Endpoints           AuthorizationServerEndpoints
	sharedErrorHandler  *auth.OAuth2ErrorHanlder
	shardTokenGranter   auth.TokenGranter
	// TODO
}

func (c *AuthorizationServerConfiguration) clientSecretEncoder() passwd.PasswordEncoder {
	if c.ClientSecretEncoder == nil {
		c.ClientSecretEncoder = passwd.NewNoopPasswordEncoder()
	}
	return c.ClientSecretEncoder
}

func (c *AuthorizationServerConfiguration) errorHandler() *auth.OAuth2ErrorHanlder {
	if c.sharedErrorHandler == nil {
		c.sharedErrorHandler = auth.NewOAuth2ErrorHanlder()
	}
	return c.sharedErrorHandler
}

func (c *AuthorizationServerConfiguration) tokenGranter() auth.TokenGranter {
	if c.shardTokenGranter == nil {
		c.shardTokenGranter = auth.NewCompositeTokenGranter(
			grants.NewAuthorizationCodeGranter(),
			grants.NewClientCredentialsGranter(),
			grants.NewPasswordGranter(),
		)
	}
	return c.shardTokenGranter
}


