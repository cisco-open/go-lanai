package init

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Precedence: 1000,
	PriorityOptions: []fx.Option{
		fx.Provide(web.BindServerProperties, gin.Default, web.NewRegistrar),
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
	Provide dependencies
***************************/


/**************************
	Setup
***************************/
type setupComponents struct {
	fx.In
	Registrar *web.Registrar
	// TODO we could include security configurations, customizations here
}

func setup(lc fx.Lifecycle, dep setupComponents) {
	lc.Append(fx.Hook{
		OnStart: makeMappingRegistrationOnStartHandler(&dep),
	})
}

func makeMappingRegistrationOnStartHandler(dep *setupComponents) bootstrap.LifecycleHandler {
	return func(ctx context.Context) (err error) {
		return dep.Registrar.Run()
	}
}