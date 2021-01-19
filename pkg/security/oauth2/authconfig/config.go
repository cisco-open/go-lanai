package authconfig

import (
	"context"
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
	ClientStore                 auth.OAuth2ClientStore
	ClientSecretEncoder         passwd.PasswordEncoder
	Endpoints                   AuthorizationServerEndpoints
	UserAccountStore            security.AccountStore
	UserPasswordEncoder         passwd.PasswordEncoder
	sharedErrorHandler          *auth.OAuth2ErrorHanlder
	sharedTokenGranter          auth.TokenGranter
	sharedTokenStore            auth.TokenStore
	sharedAuthService			auth.AuthorizationService
	sharedPasswordAuthenticator security.Authenticator
	sharedContextDetailsStore 	security.ContextDetailsStore
	// TODO
}

func (c *AuthorizationServerConfiguration) clientSecretEncoder() passwd.PasswordEncoder {
	if c.ClientSecretEncoder == nil {
		c.ClientSecretEncoder = passwd.NewNoopPasswordEncoder()
	}
	return c.ClientSecretEncoder
}

func (c *AuthorizationServerConfiguration) userPasswordEncoder() passwd.PasswordEncoder {
	if c.UserPasswordEncoder == nil {
		c.UserPasswordEncoder = passwd.NewNoopPasswordEncoder()
	}
	return c.UserPasswordEncoder
}

func (c *AuthorizationServerConfiguration) errorHandler() *auth.OAuth2ErrorHanlder {
	if c.sharedErrorHandler == nil {
		c.sharedErrorHandler = auth.NewOAuth2ErrorHanlder()
	}
	return c.sharedErrorHandler
}

func (c *AuthorizationServerConfiguration) tokenGranter() auth.TokenGranter {
	if c.sharedTokenGranter == nil {
		granters := []auth.TokenGranter {
			grants.NewAuthorizationCodeGranter(c.authorizationService()),
			grants.NewClientCredentialsGranter(c.authorizationService()),
		}

		// password granter is optional
		if c.passwordGrantAuthenticator() != nil {
			passwordGranter := grants.NewPasswordGranter(c.passwordGrantAuthenticator(), c.authorizationService())
			granters = append(granters, passwordGranter)
		}

		c.sharedTokenGranter = auth.NewCompositeTokenGranter(granters...)
	}
	return c.sharedTokenGranter
}

func (c *AuthorizationServerConfiguration) passwordGrantAuthenticator() security.Authenticator {
	if c.sharedPasswordAuthenticator == nil && c.UserAccountStore != nil {
		authenticator, err := passwd.NewAuthenticatorBuilder(
			passwd.New().
				AccountStore(c.UserAccountStore).
				PasswordEncoder(c.userPasswordEncoder()).
				MFA(false),
		).Build(context.Background())

		if err == nil {
			c.sharedPasswordAuthenticator = authenticator
		}
	}
	return c.sharedPasswordAuthenticator
}

func (c *AuthorizationServerConfiguration) contextDetailsStore() security.ContextDetailsStore {
	// TODO
	if c.sharedContextDetailsStore == nil {

	}
	//return c.sharedContextDetailsStore
	return nil
}

func (c *AuthorizationServerConfiguration) tokenStore() auth.TokenStore {
	if c.sharedTokenStore == nil {
		// TODO
		c.sharedTokenStore = auth.NewJwtTokenStore(c.contextDetailsStore())
	}
	return c.sharedTokenStore
}

func (c *AuthorizationServerConfiguration) authorizationService() auth.AuthorizationService {
	if c.sharedAuthService == nil {
		c.sharedAuthService = auth.NewDetailsPersistentAuthorizationService(func(conf *auth.AuthServiceConfig) {
			conf.TokenStore = c.tokenStore()
		})
	}

	return c.sharedAuthService
}


