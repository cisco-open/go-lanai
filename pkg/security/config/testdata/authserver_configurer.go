package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/extsamlidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/passwdidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/unknownIdp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/samltest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"go.uber.org/fx"
)

const (
	IdpDomainPasswd    = "passwd.lanai.com"
	IdpDomainExtSAML   = "saml.lanai.com"
	ExtSamlIdpName     = "ext-saml-idp"
	ExtSamlIdpEntityID = "http://external.saml.com/samlidp/metadata"
	ExtSamlIdpSSOUrl = "http://external.saml.com/samlidp/authorize"
	ExtSamlIdpSLOUrl = "http://external.saml.com/samlidp/logout"
)

type authDI struct {
	fx.In
	MockingProperties   MockingProperties
	IdpManager          idp.IdentityProviderManager
	AccountStore        security.AccountStore
	PasswordEncoder     passwd.PasswordEncoder
	Properties          authserver.AuthServerProperties
	PasswdIDPProperties passwdidp.PwdAuthProperties
	SamlIDPProperties   extsamlidp.SamlAuthProperties
}

func NewAuthServerConfigurer(di authDI) authserver.AuthorizationServerConfigurer {
	return func(config *authserver.Configuration) {
		// setup IDPs
		config.AddIdp(passwdidp.NewPasswordIdpSecurityConfigurer(
			passwdidp.WithProperties(&di.PasswdIDPProperties),
			passwdidp.WithMFAListeners(),
		))
		config.AddIdp(extsamlidp.NewSamlIdpSecurityConfigurer(
			extsamlidp.WithProperties(&di.SamlIDPProperties),
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

func NewMockedIDPManager() *samltest.MockedIdpManager {
	return samltest.NewMockedIdpManager(func(opt *samltest.IdpManagerMockOption) {
		opt.IDPList = []idp.IdentityProvider {
			extsamlidp.NewIdentityProvider(func(opt *extsamlidp.SamlIdpDetails) {
				opt.EntityId = ExtSamlIdpEntityID
				opt.Domain = IdpDomainExtSAML
				opt.ExternalIdpName = ExtSamlIdpName
				opt.ExternalIdName = "username"
				opt.MetadataLocation = "testdata/ext-saml-metadata.xml"
			}),
		}
		opt.Delegates = []idp.IdentityProviderManager{
			sectest.NewMockedIDPManager(func(opt *sectest.IdpManagerMockOption) {
				opt.PasswdIDPDomain = IdpDomainPasswd
			}),
		}
	})
}