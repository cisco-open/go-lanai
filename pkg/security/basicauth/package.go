package basicauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

//goland:noinspection GoNameStartsWithPackageName
var BasicAuthModule = &bootstrap.Module{
	Name: "basic auth",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(BasicAuthModule)
}

type initDI struct {
	fx.In
	SecRegistrar security.Registrar `optonal:true`
}

func register(init security.Registrar) {
	configurer := newBasicAuthConfigurer()
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}
