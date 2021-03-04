package authserver

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/grants"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
	"net/url"
)

type AuthorizationServerConfigurer func(*Configuration)

type authServerDI struct {
	fx.In
	AppContext           *bootstrap.ApplicationContext
	WebRegistrar         *web.Registrar
	SecurityRegistrar    security.Registrar
	Configurer           AuthorizationServerConfigurer
	RedisClientFactory   redis.ClientFactory
	SessionProperties    security.SessionProperties
	CryptoProperties     jwt.CryptoProperties
	DiscoveryCustomizers *discovery.Customizers
}

// Configuration entry point
func ConfigureAuthorizationServer(di authServerDI) {
	config := Configuration{
		appContext:         di.AppContext,
		redisClientFactory: di.RedisClientFactory,
		sessionProperties:  di.SessionProperties,
		cryptoProperties:   di.CryptoProperties,
	}
	di.Configurer(&config)

	// SMCR
	di.DiscoveryCustomizers.Add(security.CompatibilityDiscoveryCustomizer)

	// Securities
	di.SecurityRegistrar.Register(&ClientAuthEndpointsConfigurer{config: &config})
	for _, configuer := range config.idpConfigurers {
		di.SecurityRegistrar.Register(&AuthorizeEndpointConfigurer{config: &config, delegate: configuer})
	}

	// Additional endpoints
	registerEndpoints(di.WebRegistrar, &config)
}

/****************************
	configuration
 ****************************/
//TODO: constructor
type ConditionalEndpoint struct {
	Location *url.URL
	Condition web.RequestMatcher
}

type Endpoints struct {
	Authorize  ConditionalEndpoint
	Approval   string
	Token      string
	CheckToken string
	UserInfo   string
	JwkSet     string
	Logout     string
	SamlSso    ConditionalEndpoint
	SamlMetadata string
}

type Configuration struct {
	// configurable items
	ClientStore         oauth2.OAuth2ClientStore
	ClientSecretEncoder passwd.PasswordEncoder
	Endpoints           Endpoints
	UserAccountStore    security.AccountStore
	TenantStore         security.TenantStore
	ProviderStore       security.ProviderStore
	UserPasswordEncoder passwd.PasswordEncoder
	TokenStore          auth.TokenStore
	JwkStore            jwt.JwkStore

	// not directly configurable items
	appContext                *bootstrap.ApplicationContext
	redisClientFactory        redis.ClientFactory
	sessionProperties         security.SessionProperties
	cryptoProperties          jwt.CryptoProperties
	idpConfigurers            []IdpSecurityConfigurer
	sharedErrorHandler        *auth.OAuth2ErrorHandler
	sharedTokenGranter        auth.TokenGranter
	sharedAuthService         auth.AuthorizationService
	sharedPasswdAuthenticator security.Authenticator
	sharedContextDetailsStore security.ContextDetailsStore
	sharedJwtEncoder          jwt.JwtEncoder
	sharedJwtDecoder          jwt.JwtDecoder
	sharedDetailsFactory      *common.ContextDetailsFactory
	sharedARProcessor         auth.AuthorizeRequestProcessor
	sharedAuthHanlder         auth.AuthorizeHandler
	sharedAuthCodeStore       auth.AuthorizationCodeStore
	sharedTokenAuthenticator  security.Authenticator
}

func (c *Configuration) AddIdp(configurer IdpSecurityConfigurer) {
	c.idpConfigurers = append(c.idpConfigurers, configurer)
}

func (c *Configuration) clientSecretEncoder() passwd.PasswordEncoder {
	if c.ClientSecretEncoder == nil {
		c.ClientSecretEncoder = passwd.NewNoopPasswordEncoder()
	}
	return c.ClientSecretEncoder
}

func (c *Configuration) userPasswordEncoder() passwd.PasswordEncoder {
	if c.UserPasswordEncoder == nil {
		c.UserPasswordEncoder = passwd.NewNoopPasswordEncoder()
	}
	return c.UserPasswordEncoder
}

func (c *Configuration) errorHandler() *auth.OAuth2ErrorHandler {
	if c.sharedErrorHandler == nil {
		c.sharedErrorHandler = auth.NewOAuth2ErrorHanlder()
	}
	return c.sharedErrorHandler
}

func (c *Configuration) tokenGranter() auth.TokenGranter {
	if c.sharedTokenGranter == nil {
		granters := []auth.TokenGranter {
			grants.NewAuthorizationCodeGranter(c.authorizationService(), c.authorizeCodeStore()),
			grants.NewClientCredentialsGranter(c.authorizationService()),
			grants.NewRefreshGranter(c.authorizationService(), c.tokenStore()),
			grants.NewSwitchUserGranter(c.authorizationService(), c.tokenAuthenticator(), c.UserAccountStore),
			grants.NewSwitchTenantGranter(c.authorizationService(), c.tokenAuthenticator()),
		}

		// password granter is optional
		if c.passwordGrantAuthenticator() != nil {
			passwordGranter := grants.NewPasswordGranter(c.authorizationService(), c.passwordGrantAuthenticator())
			granters = append(granters, passwordGranter)
		}

		c.sharedTokenGranter = auth.NewCompositeTokenGranter(granters...)
	}
	return c.sharedTokenGranter
}

func (c *Configuration) passwordGrantAuthenticator() security.Authenticator {
	if c.sharedPasswdAuthenticator == nil && c.UserAccountStore != nil {
		authenticator, err := passwd.NewAuthenticatorBuilder(
			passwd.New().
				AccountStore(c.UserAccountStore).
				PasswordEncoder(c.userPasswordEncoder()).
				MFA(false),
		).Build(context.Background())

		if err == nil {
			c.sharedPasswdAuthenticator = authenticator
		}
	}
	return c.sharedPasswdAuthenticator
}

func (c *Configuration) contextDetailsStore() security.ContextDetailsStore {
	if c.sharedContextDetailsStore == nil {
		c.sharedContextDetailsStore = common.NewRedisContextDetailsStore(c.appContext, c.redisClientFactory)
	}
	return c.sharedContextDetailsStore
}

func (c *Configuration) authorizationRegistry() auth.AuthorizationRegistry {
	return c.contextDetailsStore().(auth.AuthorizationRegistry)
}

func (c *Configuration) tokenStore() auth.TokenStore {
	if c.TokenStore == nil {
		c.TokenStore = auth.NewJwtTokenStore(func(opt *auth.JTSOption) {
			opt.DetailsStore = c.contextDetailsStore()
			opt.Encoder = c.jwtEncoder()
			opt.Decoder = c.jwtDecoder()
			opt.AuthRegistry = c.authorizationRegistry()
			// TODO enhancers
		})
	}
	return c.TokenStore
}

func (c *Configuration) authorizationService() auth.AuthorizationService {
	if c.sharedAuthService == nil {
		c.sharedAuthService = auth.NewDefaultAuthorizationService(func(conf *auth.DASOption) {
			conf.TokenStore = c.tokenStore()
			conf.DetailsFactory = c.contextDetailsFactory()
			conf.ClientStore = c.ClientStore
			conf.AccountStore = c.UserAccountStore
			conf.TenantStore = c.TenantStore
			conf.ProviderStore = c.ProviderStore
		})
	}

	return c.sharedAuthService
}

func (c *Configuration) jwkStore() jwt.JwkStore {
	if c.JwkStore == nil {
		c.JwkStore = jwt.NewFileJwkStore(c.cryptoProperties)
	}
	return c.JwkStore
}

func (c *Configuration) jwtEncoder() jwt.JwtEncoder {
	if c.sharedJwtEncoder == nil {
		c.sharedJwtEncoder = jwt.NewRS256JwtEncoder(c.jwkStore(), c.cryptoProperties.Jwt.KeyName)
	}
	return c.sharedJwtEncoder
}

func (c *Configuration) jwtDecoder() jwt.JwtDecoder {
	if c.sharedJwtDecoder == nil {
		// TODO
		c.sharedJwtDecoder = jwt.NewRS256JwtDecoder(c.jwkStore(), c.cryptoProperties.Jwt.KeyName)
	}
	return c.sharedJwtDecoder
}

func (c *Configuration) contextDetailsFactory() *common.ContextDetailsFactory {
	if c.sharedDetailsFactory == nil {
		c.sharedDetailsFactory = common.NewContextDetailsFactory()
	}
	return c.sharedDetailsFactory
}

func (c *Configuration) authorizeRequestProcessor() auth.AuthorizeRequestProcessor {
	if c.sharedARProcessor == nil {
		//TODO OIDC extension
		std := auth.NewStandardAuthorizeRequestProcessor(func(opt *auth.StdARPOption) {
			opt.ClientStore = c.ClientStore
			opt.ResponseTypes = auth.StandardResponseTypes
		})
		c.sharedARProcessor = auth.NewCompositeAuthorizeRequestProcessor(std)
	}
	return c.sharedARProcessor
}

func (c *Configuration) authorizeHanlder() auth.AuthorizeHandler {
	if c.sharedAuthHanlder == nil {
		//TODO OIDC extension
		c.sharedAuthHanlder = auth.NewAuthorizeHandler(func(opt *auth.AuthHandlerOption) {
			//opt.Extensions = OIDC extensions
			opt.ApprovalPageTmpl = "authorize.tmpl"
			opt.ApprovalUrl = c.Endpoints.Approval
			opt.AuthService = c.authorizationService()
			opt.AuthCodeStore = c.authorizeCodeStore()
		})
	}
	return c.sharedAuthHanlder
}

func (c *Configuration) authorizeCodeStore() auth.AuthorizationCodeStore {
	if c.sharedAuthCodeStore == nil {
		c.sharedAuthCodeStore = auth.NewRedisAuthorizationCodeStore(c.appContext, c.redisClientFactory, c.sessionProperties.DbIndex)
	}
	return c.sharedAuthCodeStore
}

func (c *Configuration) tokenAuthenticator() security.Authenticator {
	if c.sharedTokenAuthenticator == nil {
		c.sharedTokenAuthenticator = tokenauth.NewAuthenticator(func(opt *tokenauth.AuthenticatorOption) {
			opt.TokenStoreReader = c.tokenStore()
		})
	}
	return c.sharedTokenAuthenticator
}
