package samlctx

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "saml",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Provide(BindSamlProperties),
	},
}

func init() {
	bootstrap.Register(Module)
}
