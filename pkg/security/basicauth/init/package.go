package init

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/bootstrap"
	"cto-github.cisco.com/livdu/jupiter/pkg/security"
	"cto-github.cisco.com/livdu/jupiter/pkg/security/basicauth"
	"cto-github.cisco.com/livdu/jupiter/pkg/web"
	"cto-github.cisco.com/livdu/jupiter/pkg/web/middleware"
	"cto-github.cisco.com/livdu/jupiter/pkg/web/route"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Provide(basicauth.NewBasicAuth),
		fx.Invoke(setup),
	},
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

/**************************
	Setup
***************************/
type setupComponents struct {
	fx.In
	BasicAuth  basicauth.Authenticator
	Registerer *web.Registrar
}

func setup(_ fx.Lifecycle, dep setupComponents) {

	auth := middleware.NewBuilder("basic auth").
		ApplyTo(route.WithPattern("/api/**")).
		Order(0).
		With(dep.BasicAuth.(web.Middleware)).
		Build()

	if err := dep.Registerer.Register(auth); err != nil {
		panic(err)
	}
}
