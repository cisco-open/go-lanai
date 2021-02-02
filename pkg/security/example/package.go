package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/authconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"go.uber.org/fx"
)

func init() {
	bootstrap.AddOptions(
		fx.Provide(BindAccountsProperties),
		fx.Provide(BindAccountPoliciesProperties),
		fx.Provide(BindClientsProperties),
		fx.Provide(NewInMemoryAccountStore),
		fx.Provide(NewInMemoryFederatedAccountStore),
		fx.Provide(NewInMemoryClientStore),
		fx.Provide(NewTenantStore),
		fx.Provide(NewProviderStore),
		fx.Provide(newAuthServerConfigurer),
		fx.Provide(NewInMemoryIdpManager),
		fx.Provide(NewInMemAuthFlowManager),
		fx.Invoke(configureSecurity),
	)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

func configureSecurity(init security.Registrar, store security.AccountStore) {
	init.Register(&TestSecurityConfigurer { accountStore: store })
	init.Register(&AnotherSecurityConfigurer { })
	init.Register(&ErrorPageSecurityConfigurer{})
}

type dependencies struct {
	fx.In
	ClientStore        oauth2.OAuth2ClientStore
	AccountStore       security.AccountStore
	TenantStore        security.TenantStore
	ProviderStore      security.ProviderStore
	RedisClientFactory redis.ClientFactory
	// TODO properties
}

func newAuthServerConfigurer(deps dependencies) authconfig.AuthorizationServerConfigurer {
	return func(config *authconfig.AuthorizationServerConfiguration) {
		config.ClientStore = deps.ClientStore
		config.ClientSecretEncoder = passwd.NewNoopPasswordEncoder()
		config.UserAccountStore = deps.AccountStore
		config.TenantStore = deps.TenantStore
		config.ProviderStore = deps.ProviderStore
		config.UserPasswordEncoder = passwd.NewNoopPasswordEncoder()
		config.JwkStore = jwt.NewStaticJwkStore("default")
		config.RedisClientFactory = deps.RedisClientFactory
		config.Endpoints = authconfig.AuthorizationServerEndpoints{
			Authorize: "/v2/authorize",
			Token: "/v2/token",
			CheckToken: "/v2/check_token",
			UserInfo: "/v2/userinfo",
			JwkSet: "/v2/jwks",
		}
	}
}
