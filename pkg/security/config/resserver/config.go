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

type resServerDI struct {
	fx.In
	AppContext           *bootstrap.ApplicationContext
	Configurer           ResourceServerConfigurer
	SecurityRegistrar    security.Registrar
	RedisClientFactory   redis.ClientFactory
	CryptoProperties     jwt.CryptoProperties
	DiscoveryCustomizers *discovery.Customizers
}

// Configuration entry point
func ConfigureResourceServer(di resServerDI) {
	config := &Configuration{
		appContext:         di.AppContext,
		cryptoProperties:   di.CryptoProperties,
		redisClientFactory: di.RedisClientFactory,
	}
	di.Configurer(config)

	// SMCR
	di.DiscoveryCustomizers.Add(security.CompatibilityDiscoveryCustomizer)

	// reigester token auth feature
	configurer := tokenauth.NewTokenAuthConfigurer(func(opt *tokenauth.TokenAuthOption) {
		opt.TokenStoreReader = config.tokenStoreReader()
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
	sharedErrorHandler        *tokenauth.OAuth2ErrorHandler
	sharedContextDetailsStore security.ContextDetailsStore
	sharedJwtDecoder          jwt.JwtDecoder
	// TODO
}

func (c *Configuration) errorHandler() *tokenauth.OAuth2ErrorHandler {
	if c.sharedErrorHandler == nil {
		c.sharedErrorHandler = tokenauth.NewOAuth2ErrorHanlder()
	}
	return c.sharedErrorHandler
}

//TODO: here we need c to have some additional properties in order to create the timeoutApplier
func (c *Configuration) contextDetailsStore() security.ContextDetailsStore {
	if c.sharedContextDetailsStore == nil {
		c.sharedContextDetailsStore = common.NewRedisContextDetailsStore(c.appContext, c.redisClientFactory)
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
