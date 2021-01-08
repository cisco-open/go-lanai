package request_cache

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"encoding/gob"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "request_cache",
	Precedence: security.MinSecurityPrecedence + 20, //after session
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(Module)

	GobRegister()
}

func GobRegister() {
	gob.Register((*CachedRequest)(nil))
}

func register(init security.Registrar) {
	configurer := newConfigurer()
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}
