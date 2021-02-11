package resserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	"go.uber.org/fx"
)

type ResourceServerConfigurer func(*Configuration)

type dependencies struct {
	fx.In
	Configurer         ResourceServerConfigurer
	SecurityRegistrar  security.Registrar
	RedisClientFactory redis.ClientFactory
	CryptoProperties   jwt.CryptoProperties
}

// Configuration entry point
func ConfigureResourceServer(deps dependencies) {
	config := &Configuration{
		redisClientFactory: deps.RedisClientFactory,
		cryptoProperties: deps.CryptoProperties,
	}
	deps.Configurer(config)

	// reigester token auth feature
	configurer := tokenauth.NewTokenAuthConfigurer(func(opt *tokenauth.TokenAuthOption) {
		opt.TokenStoreReader = config.tokenStoreReader()
	})
	deps.SecurityRegistrar.(security.FeatureRegistrar).RegisterFeature(tokenauth.FeatureId, configurer)
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

func (c *Configuration) contextDetailsStore() security.ContextDetailsStore {
	if c.sharedContextDetailsStore == nil {
		c.sharedContextDetailsStore = common.NewRedisContextDetailsStore(c.redisClientFactory)
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
		c.sharedJwtDecoder = jwt.NewRS256JwtDecoder(c.jwkStore(), "default")
	}
	return c.sharedJwtDecoder
}
