package resserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	"go.uber.org/fx"
)

type ResourceServerConfigurer func(*Configuration)

type resServerConfigDI struct {
	fx.In
	AppContext           *bootstrap.ApplicationContext
	RedisClientFactory   redis.ClientFactory
	CryptoProperties     jwt.CryptoProperties
	TimeoutSupport 	 	 oauth2.TimeoutApplier `optional:"true"`
	Configurer           ResourceServerConfigurer
}

type resServerOut struct {
	fx.Out
	Config *Configuration
	TokenStore oauth2.TokenStoreReader
}

//goland:noinspection GoExportedFuncWithUnexportedType,HttpUrlsUsage
func ProvideResServerDI(di resServerConfigDI) resServerOut {
	config := Configuration{
		appContext:         di.AppContext,
		cryptoProperties:   di.CryptoProperties,
		redisClientFactory: di.RedisClientFactory,
		timeoutSupport:     di.TimeoutSupport,
		RemoteEndpoints: RemoteEndpoints{
			Token:      "http://authserver/v2/token",
			CheckToken: "http://authserver/v2/check_token",
			UserInfo:   "http://authserver/v2/userinfo",
			JwkSet:     "http://authserver/v2/jwks",
		},
	}
	di.Configurer(&config)
	return resServerOut{
		Config: &config,
		TokenStore: config.SharedTokenStoreReader(),
	}
}

type resServerDI struct {
	fx.In
	Config               *Configuration
	SecurityRegistrar    security.Registrar
	DiscoveryCustomizers *discovery.Customizers `optional:"true"`
}

// ConfigureResourceServer configuration entry point
func ConfigureResourceServer(di resServerDI) {
	// SMCR only applicable when discovery is on
	if di.DiscoveryCustomizers != nil {
		di.DiscoveryCustomizers.Add(security.CompatibilityDiscoveryCustomizer)
	}

	// reigester token auth feature
	configurer := tokenauth.NewTokenAuthConfigurer(func(opt *tokenauth.TokenAuthOption) {
		opt.TokenStoreReader = di.Config.tokenStoreReader()
	})
	di.SecurityRegistrar.(security.FeatureRegistrar).RegisterFeature(tokenauth.FeatureId, configurer)
}

/****************************
	configuration
 ****************************/

type RemoteEndpoints struct {
	Token      string
	CheckToken string
	UserInfo   string
	JwkSet     string
}

type Configuration struct {
	// configurable items
	RemoteEndpoints  RemoteEndpoints
	TokenStoreReader oauth2.TokenStoreReader
	JwkStore         jwt.JwkStore

	// not directly configurable items
	appContext                *bootstrap.ApplicationContext
	redisClientFactory        redis.ClientFactory
	cryptoProperties          jwt.CryptoProperties
	sharedTokenAuthenticator  security.Authenticator
	sharedErrorHandler        *tokenauth.OAuth2ErrorHandler
	sharedContextDetailsStore security.ContextDetailsStore
	sharedJwtDecoder          jwt.JwtDecoder
	timeoutSupport 			  oauth2.TimeoutApplier
}

func (c *Configuration) SharedTokenStoreReader() oauth2.TokenStoreReader {
	return c.tokenStoreReader()
}

func (c *Configuration) errorHandler() *tokenauth.OAuth2ErrorHandler {
	if c.sharedErrorHandler == nil {
		c.sharedErrorHandler = tokenauth.NewOAuth2ErrorHanlder()
	}
	return c.sharedErrorHandler
}

func (c *Configuration) contextDetailsStore() security.ContextDetailsStore {
	if c.sharedContextDetailsStore == nil {
		c.sharedContextDetailsStore = common.NewRedisContextDetailsStore(c.appContext, c.redisClientFactory, c.timeoutSupport)
	}
	return c.sharedContextDetailsStore
}

func (c *Configuration) tokenStoreReader() oauth2.TokenStoreReader {
	if c.TokenStoreReader == nil {
		c.TokenStoreReader = common.NewJwtTokenStoreReader(func(opt *common.JTSROption) {
			opt.DetailsStore = c.contextDetailsStore()
			opt.Decoder = c.jwtDecoder()
		})
	}
	return c.TokenStoreReader
}

func (c *Configuration) jwkStore() jwt.JwkStore {
	if c.JwkStore == nil {
		c.JwkStore = jwt.NewFileJwkStore(c.cryptoProperties)
	}
	return c.JwkStore
}

func (c *Configuration) jwtDecoder() jwt.JwtDecoder {
	if c.sharedJwtDecoder == nil {
		c.sharedJwtDecoder = jwt.NewRS256JwtDecoder(c.jwkStore(), c.cryptoProperties.Jwt.KeyName)
	}
	return c.sharedJwtDecoder
}

func (c *Configuration) tokenAuthenticator() security.Authenticator {
	if c.sharedTokenAuthenticator == nil {
		c.sharedTokenAuthenticator = tokenauth.NewAuthenticator(func(opt *tokenauth.AuthenticatorOption) {
			opt.TokenStoreReader = c.tokenStoreReader()
		})
	}
	return c.sharedTokenAuthenticator
}