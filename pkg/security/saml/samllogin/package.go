package samllogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/gob"
	"go.uber.org/fx"
)

var logger = log.New("SAML.Auth")

var SamlAuthModule = &bootstrap.Module{
	Name: "saml authenticator",
	Precedence: security.MinSecurityPrecedence + 30,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(SamlAuthModule)

	gob.Register((*samlAssertionAuthentication)(nil))
}

type initDI struct {
	fx.In
	SecRegistrar   security.Registrar `optional:"true"`
	SamlProperties saml.SamlProperties
	ServerProps    web.ServerProperties
	IdpManager     idp.IdentityProviderManager
	AccountStore   security.FederatedAccountStore
}

func register(di initDI) {
	if di.SecRegistrar != nil {
		configurer := newSamlAuthConfigurer(di.SamlProperties, di.ServerProps, di.IdpManager, di.AccountStore)
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
	}
}