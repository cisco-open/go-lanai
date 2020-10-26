package session

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/bootstrap"
	"cto-github.cisco.com/livdu/jupiter/pkg/web"
	"cto-github.cisco.com/livdu/jupiter/pkg/web/middleware"
	"cto-github.cisco.com/livdu/jupiter/pkg/web/route"
	"go.uber.org/fx"
	"net/http"
)

var Module = &bootstrap.Module{
	Precedence: -1,
	Options: []fx.Option{
		fx.Provide(NewManager),
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
	Registerer *web.Registrar
	SessionManager *Manager
}

func setup(_ fx.Lifecycle, dep setupComponents) {
	var matcher = route.WithPrefix("/page").
		Or(route.WithRegex("/static/.*")).
		Or(route.WithPattern("/api/**"))

	session := middleware.NewBuilder("session").
		ApplyTo(matcher).
		Order(-1).
		Use(dep.SessionManager.SessionHandlerFunc()).
		WithCondition(func (r *http.Request) bool { return true }).
		Build()

	postSession := middleware.NewBuilder("post-session").
		ApplyTo(matcher).
		Order(0).
		Use(dep.SessionManager.SessionPostHandlerFunc()).
		WithCondition(func (r *http.Request) bool { return true }).
		Build()

	if err := dep.Registerer.Register(session, postSession); err != nil {
		panic(err)
	}
}
