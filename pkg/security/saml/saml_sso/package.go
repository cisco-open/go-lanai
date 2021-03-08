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

func init() {
	bootstrap.Register(Module)
}

type dependencies struct {
	fx.In
	Properties             saml.SamlProperties
	ServerProperties       web.ServerProperties
	ServiceProviderManager SamlClientStore
	AccountStore           security.AccountStore
	AttributeGenerator     AttributeGenerator `optional:"true"`
}

func register(init security.Registrar, di dependencies) {
	configurer := newSamlAuthorizeEndpointConfigurer(di.Properties, di.ServerProperties,
		di.ServiceProviderManager, di.AccountStore,
		di.AttributeGenerator)
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}