package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/resserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/passwdidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/samlidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	sdcustomizer "cto-github.cisco.com/NFV-BU/go-lanai/pkg/servicedisc/customizer"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"go.uber.org/fx"
	"net/url"
)

var logger = log.New("SEC.Example")

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
		fx.Provide(NewInMemoryIdpManager),
		fx.Provide(NewInMemAuthFlowManager),
		fx.Provide(NewInMemSpManager),
		fx.Provide(newAuthServerConfigurer),
		fx.Provide(newResServerConfigurer),
		fx.Invoke(configureSecurity, configureConsulRegistration),
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
	ClientStore      oauth2.OAuth2ClientStore
	AccountStore     security.AccountStore
	TenantStore      security.TenantStore
	ProviderStore    security.ProviderStore
	AuthFlowManager  idp.AuthFlowManager
	// TODO properties
}

func newAuthServerConfigurer(deps dependencies) authserver.AuthorizationServerConfigurer {
	return func(config *authserver.Configuration) {
		config.AddIdp(passwdidp.NewPasswordIdpSecurityConfigurer(deps.AuthFlowManager))
		config.AddIdp(samlidp.NewSamlIdpSecurityConfigurer(deps.AuthFlowManager))

		config.ClientStore = deps.ClientStore
		config.ClientSecretEncoder = passwd.NewNoopPasswordEncoder()
		config.UserAccountStore = deps.AccountStore
		config.TenantStore = deps.TenantStore
		config.ProviderStore = deps.ProviderStore
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

func newResServerConfigurer(deps dependencies) resserver.ResourceServerConfigurer {
	return func(config *resserver.Configuration) {
		config.RemoteEndpoints = resserver.RemoteEndpoints{
			Token: "http://localhost:8080/europa/v2/token",
			CheckToken: "http://localhost:8080/europa/v2/check_token",
			UserInfo: "http://localhost:8080/europa/v2/userinfo",
			JwkSet: "http://localhost:8080/europa/v2/jwks",
		}
	}
}

func configureConsulRegistration(r *sdcustomizer.Registrar) {
	r.Add(&RegistrationCusomizer{})
}