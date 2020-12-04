package csrf

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
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

func register(init security.Registrar, serverProps web.ServerProperties) {
	configurer := newCsrfConfigurer()
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}