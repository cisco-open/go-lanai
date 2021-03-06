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

type initDI struct {
	fx.In
	SecRegistrar security.Registrar `optional:"true"`
}

func register(di initDI) {
	if di.SecRegistrar != nil {
		configurer := newConfigurer()
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
	}
}
