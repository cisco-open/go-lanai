package authconfig

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/grants"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
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
	ClientStore                 oauth2.OAuth2ClientStore
	ClientSecretEncoder         passwd.PasswordEncoder
	Endpoints                   AuthorizationServerEndpoints
	UserAccountStore            security.AccountStore
	TenantStore                 security.TenantStore
	ProviderStore               security.ProviderStore
	UserPasswordEncoder         passwd.PasswordEncoder
	TokenStore                  auth.TokenStore
	JwkStore                    jwt.JwkStore
	sharedErrorHandler          *auth.OAuth2ErrorHanlder
	sharedTokenGranter          auth.TokenGranter
	sharedAuthService           auth.AuthorizationService
	sharedPasswordAuthenticator security.Authenticator
	sharedContextDetailsStore   security.ContextDetailsStore
	sharedJwtEncoder            jwt.JwtEncoder
	sharedJwtDecoder            jwt.JwtDecoder
	sharedDetailsFactory        *common.ContextDetailsFactory
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
	if c.TokenStore == nil {
		c.TokenStore = auth.NewJwtTokenStore(func(opt *auth.JTSOption) {
			opt.DetailsStore = c.contextDetailsStore()
			opt.Encoder = c.jwtEncoder()
			opt.Decoder = c.jwtDecoder()
			// TODO enhancers
		})
	}
	return c.TokenStore
}

func (c *AuthorizationServerConfiguration) authorizationService() auth.AuthorizationService {
	if c.sharedAuthService == nil {
		c.sharedAuthService = auth.NewDefaultAuthorizationService(func(conf *auth.DASOption) {
			conf.TokenStore = c.tokenStore()
			conf.DetailsFactory = c.contextDetailsFactory()
		})
	}

	return c.sharedAuthService
}

func (c *AuthorizationServerConfiguration) jwkStore() jwt.JwkStore {
	if c.JwkStore == nil {
		// TODO
		c.JwkStore = jwt.NewStaticJwkStore("default")
	}
	return c.JwkStore
}

func (c *AuthorizationServerConfiguration) jwtEncoder() jwt.JwtEncoder {
	if c.sharedJwtEncoder == nil {
		// TODO
		c.sharedJwtEncoder = jwt.NewRS256JwtEncoder(c.jwkStore(), "default")
	}
	return c.sharedJwtEncoder
}

func (c *AuthorizationServerConfiguration) jwtDecoder() jwt.JwtDecoder {
	if c.sharedJwtDecoder == nil {
		// TODO
		c.sharedJwtDecoder = jwt.NewRS256JwtDecoder(c.jwkStore(), "default")
	}
	return c.sharedJwtDecoder
}

func (c *AuthorizationServerConfiguration) contextDetailsFactory() *common.ContextDetailsFactory {
	if c.sharedDetailsFactory == nil {
		c.sharedDetailsFactory = common.NewContextDetailsFactory(func(opt *common.FactoryOption) {
			opt.ClientStore = c.ClientStore
			opt.AccountStore = c.UserAccountStore
			opt.TenantStore = c.TenantStore
			opt.ProviderStore = c.ProviderStore
		})
	}
	return c.sharedDetailsFactory
}


