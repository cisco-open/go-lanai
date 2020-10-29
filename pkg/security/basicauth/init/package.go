package init

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/bootstrap"
	"cto-github.cisco.com/livdu/jupiter/pkg/security"
	"cto-github.cisco.com/livdu/jupiter/pkg/security/basicauth"
	"cto-github.cisco.com/livdu/jupiter/pkg/security/passwd"
	"cto-github.cisco.com/livdu/jupiter/pkg/web"
	"cto-github.cisco.com/livdu/jupiter/pkg/web/middleware"
	"cto-github.cisco.com/livdu/jupiter/pkg/web/route"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Provide(basicauth.NewBasicAuthMiddleware, newAuthenticator),
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
	Providers
***************************/
// TODO this part should be moved
type passwdAuthDependencies struct {
	fx.In
	Store security.AccountStore
	PasswdEncoder passwd.PasswordEncoder `optional:"true"`
}

func newAuthenticator(d passwdAuthDependencies) security.Authenticator {
	// TODO use configurer
	passwdAuth := passwd.NewAuthenticator(d.Store, d.PasswdEncoder)
	return security.NewAuthenticator(passwdAuth)
}

/**************************
	Setup
***************************/
type setupComponents struct {
	fx.In
	BasicAuth  *basicauth.BasicAuthMiddleware
	Registerer *web.Registrar
}

func setup(_ fx.Lifecycle, dep setupComponents) {

	auth := middleware.NewBuilder("basic auth").
		ApplyTo(route.WithPattern("/api/**")).
		Order(0).
		With(web.Middleware(dep.BasicAuth)).
		Build()

	if err := dep.Registerer.Register(auth); err != nil {
		panic(err)
	}
}
