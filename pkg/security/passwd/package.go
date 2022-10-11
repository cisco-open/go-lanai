package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

var logger = log.New("SEC.Passwd")

var Module = &bootstrap.Module{
	Name: "passwd authenticator",
	Precedence: security.MinSecurityPrecedence + 30,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(Module)
}

type initDI struct {
	fx.In
	SecRegistrar    security.Registrar    `optional:"true"`
	AccountStore    security.AccountStore `optional:"true"`
	PasswordEncoder PasswordEncoder       `optional:"true"`
	Redis           redis.Client          `optional:"true"`
}

func register(di initDI) {
	if di.SecRegistrar != nil {
		configurer := newPasswordAuthConfigurer(di.AccountStore, di.PasswordEncoder, di.Redis)
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(PasswordAuthenticatorFeatureId, configurer)
	}
}