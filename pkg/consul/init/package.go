package consul

import (
	"embed"

	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	appconfigInit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	consulhealth "cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul/health"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

//go:embed defaults-consul.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name:       "consul",
	Precedence: bootstrap.ConsulPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(BindConnectionProperties),
		fx.Provide(ProvideDefaultClient),
	},
	Options: []fx.Option{
		appconfigInit.FxEmbeddedDefaults(defaultConfigFS),
		fx.Invoke(consulhealth.Register),
	},
}

func Use() {
	bootstrap.Register(Module)
}

func BindConnectionProperties(bootstrapConfig *appconfig.BootstrapConfig) consul.ConnectionProperties {
	c := consul.ConnectionProperties{}
	if e := bootstrapConfig.Bind(&c, consul.PropertyPrefix); e != nil {
		panic(errors.Wrap(e, "failed to bind consul's ConnectionProperties"))
	}
	return c
}

type clientDI struct {
	fx.In
	Props       consul.ConnectionProperties
	Customizers []consul.Options `group:"consul"`
}

func ProvideDefaultClient(di clientDI) (*consul.Connection, error) {
	opts := append([]consul.Options{consul.WithProperties(di.Props)}, di.Customizers...)
	return consul.New(opts...)
}
