package token

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

//goland:noinspection GoNameStartsWithPackageName
var Module = &bootstrap.Module{
	Name: "oauth2 auth - token",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(Module)
}

func register(init security.Registrar) {
	configurer := newOAuth2AuthConfigurer()
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}
