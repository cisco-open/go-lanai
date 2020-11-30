package access

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

//goland:noinspection GoNameStartsWithPackageName
var AccessControlModule = &bootstrap.Module{
	Name: "access control",
	Precedence: security.MinSecurityPrecedence + 30,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(AccessControlModule)
}

func register(init security.Registrar) {
	configurer := newAccessControlConfigurer()
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}
