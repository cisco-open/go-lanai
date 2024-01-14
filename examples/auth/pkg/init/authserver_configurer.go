package serviceinit

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/extsamlidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/passwdidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/unknownIdp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"go.uber.org/fx"
)

type authDI struct {
	fx.In
	ClientStore         oauth2.OAuth2ClientStore
	AccountStore        security.AccountStore
	TenantStore         security.TenantStore
	ProviderStore       security.ProviderStore
	IdpManager          idp.IdentityProviderManager
	PasswdIDPProperties passwdidp.PwdAuthProperties
}

func newAuthServerConfigurer(di authDI) authserver.AuthorizationServerConfigurer {
	return func(config *authserver.Configuration) {
		config.AddIdp(passwdidp.NewPasswordIdpSecurityConfigurer(passwdidp.WithProperties(&di.PasswdIDPProperties)))
		config.AddIdp(extsamlidp.NewSamlIdpSecurityConfigurer())
		config.AddIdp(unknownIdp.NewNoIdpSecurityConfigurer())

		config.IdpManager = di.IdpManager
		config.ClientStore = di.ClientStore
		config.ClientSecretEncoder = passwd.NewNoopPasswordEncoder()
		config.UserAccountStore = di.AccountStore
		config.TenantStore = di.TenantStore
		config.ProviderStore = di.ProviderStore
		config.UserPasswordEncoder = passwd.NewNoopPasswordEncoder()
	}
}
