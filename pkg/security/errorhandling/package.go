package errorhandling

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

var logger = log.New("SEC.Err")

//goland:noinspection GoNameStartsWithPackageName
var ErrorHandlingModule = &bootstrap.Module{
	Name: "basic auth",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(ErrorHandlingModule)
}

func register(init security.Registrar) {
	configurer := newErrorHandlingConfigurer()
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}
