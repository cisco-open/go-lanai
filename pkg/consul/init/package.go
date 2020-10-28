package init

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/bootstrap"
	"cto-github.cisco.com/livdu/jupiter/pkg/appconfig"
	"cto-github.cisco.com/livdu/jupiter/pkg/consul"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Precedence: -1,
	PriorityOptions: []fx.Option{
		fx.Provide(newConnectionProperties),
		fx.Provide(newConsulConnection),
	},
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

type bootstrapConfigParam struct {
	fx.In
	Config *appconfig.Config `name:"bootstrap_config"`
}

func newConnectionProperties(param bootstrapConfigParam) *consul.ConnectionProperties {
	c := &consul.ConnectionProperties{
		//TODO: defaults can be specified here
	}
	param.Config.Bind(c, "spring.cloud.consul")
	return c
}

func newConsulConnection(connectionProperties *consul.ConnectionProperties) *consul.Connection {
	connection, _ := consul.NewConnection(connectionProperties)
	return connection
}