package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "security",
	Precedence: MaxSecurityPrecedence,
	Options: []fx.Option{
		fx.Provide(New, newGlobalAuthenticator),
		fx.Invoke(initialize),
	},
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

/**************************
	Provider
***************************/
func newGlobalAuthenticator() Authenticator {
	return NewAuthenticator()
}

type dependencies struct {
	fx.In
	Registerer *web.Registrar
	Initializer Initializer
}

func initialize(di dependencies) {
	// TODO error handling
	if err := di.Initializer.(*initializer).initialize(di.Registerer); err != nil {
		//panic(err)
	}
}

