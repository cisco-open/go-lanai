package csrf

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "csrf",
	Precedence: security.MinSecurityPrecedence + 20, //after session
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(Module)
}

type initDI struct {
	fx.In
	SecRegistrar security.Registrar `optional:"true"`
}

func register(di initDI) {
	if di.SecRegistrar != nil {
		configurer := newCsrfConfigurer()
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
	}
}