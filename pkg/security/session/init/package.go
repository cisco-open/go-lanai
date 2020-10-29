package init

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/bootstrap"
	"cto-github.cisco.com/livdu/jupiter/pkg/security"
	"cto-github.cisco.com/livdu/jupiter/pkg/security/session"
	"cto-github.cisco.com/livdu/jupiter/pkg/security/session/store"
	"cto-github.cisco.com/livdu/jupiter/pkg/web"
	"cto-github.cisco.com/livdu/jupiter/pkg/web/middleware"
	"cto-github.cisco.com/livdu/jupiter/pkg/web/route"
	"fmt"
	"go.uber.org/fx"
	"net/http"
)

var Module = &bootstrap.Module{
	Precedence: security.MinSecurityPrecedence + 10,
	Options: []fx.Option{
		fx.Provide(session.NewManager, newSessionStore, security.BindSessionProperties),
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
	Provider
***************************/
func newSessionStore(properties security.SessionProperties) session.Store {
	secret := []byte(properties.Secret)
	switch properties.StoreType {
	case security.SessionStoreTypeMemory:
		return store.NewMemoryStore(secret)
	default:
		panic(fmt.Errorf("unsupported session storage"))
	}
}

/**************************
	Setup
***************************/
type setupComponents struct {
	fx.In
	Registerer *web.Registrar
	SessionManager *session.Manager
}

func setup(_ fx.Lifecycle, dep setupComponents) {
	var matcher = route.WithPrefix("/page").
		Or(route.WithRegex("/static/.*")).
		Or(route.WithPattern("/api/**"))

	sessionMiddleware := middleware.NewBuilder("sessionMiddleware").
		ApplyTo(matcher).
		Order(-1).
		Use(dep.SessionManager.SessionHandlerFunc()).
		WithCondition(func (r *http.Request) bool { return true }).
		Build()

	sessionTestMiddleware := middleware.NewBuilder("post-sessionMiddleware").
		ApplyTo(matcher).
		Order(0).
		Use(session.SessionDebugHandlerFunc()).
		WithCondition(func (r *http.Request) bool { return true }).
		Build()

	if err := dep.Registerer.Register(sessionMiddleware, sessionTestMiddleware); err != nil {
		panic(err)
	}
}
