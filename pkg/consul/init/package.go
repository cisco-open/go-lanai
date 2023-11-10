package consul

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name:       "consul",
	Precedence: bootstrap.ConsulPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(newConnectionProperties),
		fx.Provide(newConsulConnection),
	},
	Options: []fx.Option{
		fx.Invoke(registerHealth),
	},
}

func init() {
	bootstrap.Register(Module)
}

func Use() {
	// does nothing. Allow service to include this module in main()
}

func newConnectionProperties(bootstrapConfig *appconfig.BootstrapConfig) *consul.ConnectionProperties {
	c := &consul.ConnectionProperties{}
	if e := bootstrapConfig.Bind(c, consul.PropertyPrefix); e != nil {
		panic(errors.Wrap(e, "failed to bind ConnectionProperties"))
	}
	return c
}

func newConsulConnection(connectionProperties *consul.ConnectionProperties) *consul.Connection {
	connection, _ := consul.NewConnection(connectionProperties)
	return connection
}

type regDI struct {
	fx.In
	HealthRegistrar  health.Registrar `optional:"true"`
	ConsulConnection *consul.Connection
}

func registerHealth(di regDI) {
	if di.HealthRegistrar == nil {
		return
	}
	di.HealthRegistrar.MustRegister(consul.NewConsulHealthIndicator(di.ConsulConnection))
}
