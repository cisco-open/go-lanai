package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

var PasswordAuthModule = &bootstrap.Module{
	Name: "passwd authenticator",
	Precedence: security.MinSecurityPrecedence + 30,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(PasswordAuthModule)
}

type dependencies struct {
	fx.In
	AccountStore security.AccountStore `optional:"true"`
	PasswordEncoder PasswordEncoder `optional:"true"`
}

func register(init security.Registrar, di dependencies) {
	configurer := newPasswordAuthConfigurer(di.AccountStore, di.PasswordEncoder)
	init.(security.FeatureRegistrar).RegisterFeature(PasswordAuthenticatorFeatureId, configurer)
}