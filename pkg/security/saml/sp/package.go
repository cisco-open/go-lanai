package sp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/gob"
	"go.uber.org/fx"
)

var logger = log.New("SAML.Auth")

var Module = &bootstrap.Module{
	Name: "saml authenticator",
	Precedence: security.MinSecurityPrecedence + 30,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	gob.Register((*samlAssertionAuthentication)(nil))
}

func Use() {
	bootstrap.Register(Module)
}

type initDI struct {
	fx.In
	SecRegistrar   security.Registrar `optional:"true"`
	SamlProperties samlctx.SamlProperties
	ServerProps    web.ServerProperties
	IdpManager     idp.IdentityProviderManager
	AccountStore   security.FederatedAccountStore
}

func register(di initDI) {
	if di.SecRegistrar != nil {
		shared := newSamlConfigurer(di.SamlProperties, di.IdpManager)
		loginConfigurer := newSamlAuthConfigurer(shared, di.AccountStore)
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, loginConfigurer)

		logoutConfigurer := newSamlLogoutConfigurer(shared)
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(LogoutFeatureId, logoutConfigurer)
	}
}