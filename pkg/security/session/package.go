package session

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/bootstrap"
	"cto-github.cisco.com/livdu/jupiter/pkg/web"
	"cto-github.cisco.com/livdu/jupiter/pkg/web/middleware"
	"cto-github.cisco.com/livdu/jupiter/pkg/web/route"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Precedence: -1,
	Provides: []fx.Option{fx.Provide(NewManager)},
	Invokes: []fx.Option{fx.Invoke(setup)},
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
	session := middleware.NewBuilder("basic auth").
		ApplyTo(route.WithPrefix("/page").Or(route.WithRegex("/static/.*")) ).
		Use(dep.SessionManager.SessionHandlerFunc()).
		Build()

	if err := dep.Registerer.Register(session); err != nil {
		panic(err)
	}
}
