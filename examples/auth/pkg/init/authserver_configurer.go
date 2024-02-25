package serviceinit

import (
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/config/authserver"
	"github.com/cisco-open/go-lanai/pkg/security/idp"
	"github.com/cisco-open/go-lanai/pkg/security/idp/extsamlidp"
	"github.com/cisco-open/go-lanai/pkg/security/idp/passwdidp"
	"github.com/cisco-open/go-lanai/pkg/security/idp/unknownIdp"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/passwd"
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
