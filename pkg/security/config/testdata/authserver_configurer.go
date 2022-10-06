package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/passwdidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/samlidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/unknownIdp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"go.uber.org/fx"
)

type authDI struct {
	fx.In
	MockingProperties   MockingProperties
	IdpManager          idp.IdentityProviderManager
	AccountStore        security.AccountStore
	PasswordEncoder     passwd.PasswordEncoder
	Properties          authserver.AuthServerProperties
	PasswdIDPProperties passwdidp.PwdAuthProperties
	SamlIDPProperties   samlidp.SamlAuthProperties
}

func NewAuthServerConfigurer(di authDI) authserver.AuthorizationServerConfigurer {
	return func(config *authserver.Configuration) {
		// setup IDPs
		config.AddIdp(passwdidp.NewPasswordIdpSecurityConfigurer(
			passwdidp.WithProperties(&di.PasswdIDPProperties),
			passwdidp.WithMFAListeners(),
		))
		config.AddIdp(samlidp.NewSamlIdpSecurityConfigurer(
			samlidp.WithProperties(&di.SamlIDPProperties),
		))
		config.AddIdp(unknownIdp.NewNoIdpSecurityConfigurer())

		config.IdpManager = di.IdpManager
		config.ClientStore = sectest.NewMockedClientStore(MapValues(di.MockingProperties.Clients)...)
		config.ClientSecretEncoder = di.PasswordEncoder
		config.UserAccountStore = di.AccountStore
		config.TenantStore = sectest.NewMockedTenantStore(MapValues(di.MockingProperties.Tenants)...)
		config.ProviderStore = sectest.MockedProviderStore{}
		config.UserPasswordEncoder = di.PasswordEncoder
		config.SessionSettingService = StaticSessionSettingService(1)
	}
}

type StaticSessionSettingService int

func (s StaticSessionSettingService) GetMaximumSessions(ctx context.Context) int {
	return int(s)
}
