package init

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/bootstrap"
	"cto-github.cisco.com/livdu/jupiter/pkg/security"
	"cto-github.cisco.com/livdu/jupiter/pkg/security/passwd"
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

const (
	MWOrderSessionHandling = security.HighestMiddlewareOrder + 100
	MWOrderAuthPersistence = MWOrderSessionHandling + 10
)

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
		registerTypes()
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

	sessionHandler := middleware.NewBuilder("sessionMiddleware").
		ApplyTo(matcher).
		Order(MWOrderSessionHandling).
		Use(dep.SessionManager.SessionHandlerFunc()).
		WithCondition(func (r *http.Request) bool { return true }).
		Build()

	authPersist := middleware.NewBuilder("sessionMiddleware").
		ApplyTo(matcher).
		Order(MWOrderAuthPersistence).
		Use(dep.SessionManager.AuthenticationPersistenceHandlerFunc()).
		WithCondition(func (r *http.Request) bool { return true }).
		Build()

	test := middleware.NewBuilder("post-sessionMiddleware").
		ApplyTo(matcher).
		Order(MWOrderAuthPersistence + 10).
		Use(session.SessionDebugHandlerFunc()).
		WithCondition(func (r *http.Request) bool { return true }).
		Build()

	if err := dep.Registerer.Register(sessionHandler, authPersist, test); err != nil {
		panic(err)
	}
}

func registerTypes() {
	security.GobRegister()
	passwd.GobRegister()
}
