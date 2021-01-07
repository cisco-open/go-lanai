package formlogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
)

//goland:noinspection GoNameStartsWithPackageName
var Module = &bootstrap.Module{
	Name: "form login",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(Module)
}

func register(init security.Registrar, sessionProps security.SessionProperties, serverProps web.ServerProperties) {
	configurer := newFormLoginConfigurer(sessionProps.Cookie, serverProps)
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}
