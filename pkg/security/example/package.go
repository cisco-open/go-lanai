package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/resserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/passwdidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/samlidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"go.uber.org/fx"
	"net/url"
)

var logger = log.New("SEC.Example")

func init() {
	authserver.Use()
	resserver.Use()
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {
	bootstrap.AddOptions(
		fx.Provide(BindAccountsProperties),
		fx.Provide(BindAccountPoliciesProperties),
		fx.Provide(BindClientsProperties),
		fx.Provide(NewInMemoryAccountStore),
		fx.Provide(NewInMemoryFederatedAccountStore),
		fx.Provide(NewInMemoryClientStore),
		fx.Provide(NewTenantStore),
		fx.Provide(NewProviderStore),
		fx.Provide(NewInMemoryIdpManager),
		fx.Provide(NewInMemSpManager),
		fx.Provide(newAuthServerConfigurer),
		fx.Provide(newResServerConfigurer),
		fx.Invoke(configureSecurity),
		fx.Invoke(configureConsulRegistration),
	)
}

func configureSecurity(init security.Registrar, store security.AccountStore) {
	init.Register(&TestSecurityConfigurer { accountStore: store })
	init.Register(&AnotherSecurityConfigurer { })
	init.Register(&ErrorPageSecurityConfigurer{})
}

type authDI struct {
	fx.In
	ClientStore   oauth2.OAuth2ClientStore
	AccountStore  security.AccountStore
	TenantStore   security.TenantStore
	ProviderStore security.ProviderStore
	IdpManager    idp.IdentityProviderManager
	// TODO properties
}

func newAuthServerConfigurer(di authDI) authserver.AuthorizationServerConfigurer {
	return func(config *authserver.Configuration) {
		config.AddIdp(passwdidp.NewPasswordIdpSecurityConfigurer(di.IdpManager))
		config.AddIdp(samlidp.NewSamlIdpSecurityConfigurer(di.IdpManager))

		config.ClientStore = di.ClientStore
		config.ClientSecretEncoder = passwd.NewNoopPasswordEncoder()
		config.UserAccountStore = di.AccountStore
		config.TenantStore = di.TenantStore
		config.ProviderStore = di.ProviderStore
		config.UserPasswordEncoder = passwd.NewNoopPasswordEncoder()
		config.Endpoints = authserver.Endpoints{
			Authorize: authserver.ConditionalEndpoint{
				Location: &url.URL{Path: "/v2/authorize"},
				Condition: matcher.NotRequest(matcher.RequestWithParam("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")),
			},
			Approval: "/v2/approve",
			Token: "/v2/token",
			CheckToken: "/v2/check_token",
			UserInfo: "/v2/userinfo",
			JwkSet: "/v2/jwks",
			Logout: "/v2/logout",
			SamlSso: authserver.ConditionalEndpoint{
				Location: &url.URL{Path:"/v2/authorize", RawQuery: "grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer"},
				Condition: matcher.RequestWithParam("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer"),
			},
			SamlMetadata: "/metadata",
		}
	}
}

func newResServerConfigurer(deps authDI) resserver.ResourceServerConfigurer {
	return func(config *resserver.Configuration) {
		config.RemoteEndpoints = resserver.RemoteEndpoints{
			Token: "http://localhost:8080/europa/v2/token",
			CheckToken: "http://localhost:8080/europa/v2/check_token",
			UserInfo: "http://localhost:8080/europa/v2/userinfo",
			JwkSet: "http://localhost:8080/europa/v2/jwks",
		}
	}
}

func configureConsulRegistration(r *discovery.Customizers) {
	r.Add(&RegistrationCustomizer{})
}