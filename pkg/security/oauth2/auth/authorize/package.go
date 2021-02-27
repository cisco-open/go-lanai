package authorize

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

//goland:noinspection GoNameStartsWithPackageName
var Module = &bootstrap.Module{
	Name: "oauth2 auth - authorize",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

var logger = log.New("AuthorizeEndpoint")

func init() {
	bootstrap.Register(Module)
}

func register(init security.Registrar) {
	configurer := newOAuth2AuhtorizeEndpointConfigurer()
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}
