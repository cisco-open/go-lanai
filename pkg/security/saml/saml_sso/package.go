package saml_auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "saml auth - authorize",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

var logger = log.New("SAML.SSO")

func Use() {
	bootstrap.Register(Module)
}

type initDI struct {
	fx.In
	SecRegistrar           security.Registrar `optional:"true"`
	Properties             saml.SamlProperties
	ServerProperties       web.ServerProperties
	ServiceProviderManager SamlClientStore `optional:"true"`
	AccountStore           security.AccountStore `optional:"true"`
	AttributeGenerator     AttributeGenerator `optional:"true"`
}

func register(di initDI) {
	if di.SecRegistrar != nil {
		configurer := newSamlAuthorizeEndpointConfigurer(di.Properties,
			di.ServiceProviderManager, di.AccountStore,
			di.AttributeGenerator)
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
	}
}