package web

import (
	"context"
	"cto-github.cisco.com/livdu/jupiter/pkg/bootstrap"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Precedence: 0,
	Provides: []fx.Option{fx.Provide(gin.Default, web)},
	Invokes: []fx.Option{fx.Invoke(setup)},
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

/**************************
	Provide dependencies
***************************/
type prerequisites struct {
	fx.In
	GinEngine *gin.Engine
}

type components struct {
	fx.Out
	Registrar *Registrar
}

func web(d prerequisites) (components, error) {
	return components{
		Registrar: NewRegistrar(d.GinEngine),
	}, nil
}

/**************************
	Setup
***************************/
type setupComponents struct {
	fx.In
	Registrar *Registrar
	GinEngine *gin.Engine
	// TODO we could include security configurations, customizations here
}
func setup(lc fx.Lifecycle, dep setupComponents) {
	lc.Append(fx.Hook{
		OnStart: makeMappingRegistrationOnStartHandler(&dep),
	})
}

func makeMappingRegistrationOnStartHandler(dep *setupComponents) bootstrap.LifecycleHandler {
	return func(ctx context.Context) (err error) {
		err = dep.Registrar.initialize()
		errChan := make(chan error)
		go func() {
			errChan <- dep.GinEngine.Run()
			defer func() {close(errChan)}()
		}()
		return
	}
}