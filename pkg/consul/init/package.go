package init

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
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

func newConnectionProperties(bootstrapConfig *appconfig.BootstrapConfig) *consul.ConnectionProperties {
	c := &consul.ConnectionProperties{
		//TODO: defaults can be specified here
	}
	bootstrapConfig.Bind(c, consul.ConfigRootConsulConnection)
	return c
}

func newConsulConnection(connectionProperties *consul.ConnectionProperties) *consul.Connection {
	connection, _ := consul.NewConnection(connectionProperties)
	return connection
}