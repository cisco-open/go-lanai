package init

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/bootstrap"
	"cto-github.cisco.com/livdu/jupiter/pkg/security"
	"cto-github.cisco.com/livdu/jupiter/pkg/security/session"
	"cto-github.cisco.com/livdu/jupiter/pkg/security/session/store"
	"cto-github.cisco.com/livdu/jupiter/pkg/web"
	"cto-github.cisco.com/livdu/jupiter/pkg/web/middleware"
	"cto-github.cisco.com/livdu/jupiter/pkg/web/route"
	"go.uber.org/fx"
	"net/http"
)

var Module = &bootstrap.Module{
	Precedence: security.MinSecurityPrecedence + 10,
	Options: []fx.Option{
		fx.Provide(session.NewManager, NewSessionStore),
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
func NewSessionStore(ctx *bootstrap.ApplicationContext) session.Store {
	var secret []byte
	switch v := ctx.Value("security.session.secret"); v.(type) {
	case string:
		secret = []byte(v.(string))
	default:
		return nil
	}

	// TODO create different type based on properties
	return store.NewMemoryStore(secret)
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
		Use(dep.SessionManager.SessionTestHandlerFunc()).
		WithCondition(func (r *http.Request) bool { return true }).
		Build()

	if err := dep.Registerer.Register(sessionMiddleware, sessionTestMiddleware); err != nil {
		panic(err)
	}
}
