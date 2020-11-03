package init

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/route"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Provide(basicauth.NewBasicAuthMiddleware, newAuthenticator),
		fx.Invoke(setup),
	},
}

const (
	MWOrderBasicAuth = security.HighestMiddlewareOrder + 200
)

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
		Order(MWOrderBasicAuth).
		With(web.Middleware(dep.BasicAuth)).
		Build()

	if err := dep.Registerer.Register(auth); err != nil {
		panic(err)
	}
}
