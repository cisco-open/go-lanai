package init

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/cors"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "web",
	Precedence: bootstrap.FrameworkModulePrecedence + 1000,
	PriorityOptions: []fx.Option{
		fx.Provide(web.BindServerProperties,
			web.NewEngine,
			web.NewRegistrar),
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
	Registrar      *web.Registrar
	CorsCustomizer *cors.Customizer
	// TODO we could include security configurations, customizations here
}

func setup(lc fx.Lifecycle, dep setupComponents) {
	dep.Registrar.Register(dep.CorsCustomizer)
	lc.Append(fx.Hook{
		OnStart: makeMappingRegistrationOnStartHandler(&dep),
	})
}

func makeMappingRegistrationOnStartHandler(dep *setupComponents) bootstrap.LifecycleHandler {
	return func(ctx context.Context) (err error) {
		return dep.Registrar.Run()
	}
}