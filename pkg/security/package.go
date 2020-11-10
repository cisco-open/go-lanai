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
		fx.Provide(provideSecurityInitialization),
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
type global struct {
	fx.Out
	Initializer Initializer
	Registrar Registrar
	Authenticator Authenticator
}

// We let configurer.initializer can be autowired as both Initializer and Registrar
func provideSecurityInitialization() global {
	initializer := newSecurity()
	return global{
		Initializer: initializer,
		Registrar: initializer,
		Authenticator: NewAuthenticator(),
	}
}

/**************************
	Initialize
***************************/
type dependencies struct {
	fx.In
	Registerer  *web.Registrar
	Initializer Initializer
}

func initialize(di dependencies) {
	// TODO error handling
	if err := di.Initializer.Initialize(di.Registerer); err != nil {
		//panic(err)
	}
}

