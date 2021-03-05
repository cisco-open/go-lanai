package access

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

var logger = log.New("SEC.Access")

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

type initDI struct {
	fx.In
	SecRegistrar security.Registrar `optonal:true`
}

func register(di initDI) {
	if di.SecRegistrar != nil {
		configurer := newAccessControlConfigurer()
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
	}
}
