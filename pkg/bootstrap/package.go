package bootstrap

import (
	"context"
	"cto-github.cisco.com/livdu/jupiter/pkg/config"
	"cto-github.cisco.com/livdu/jupiter/pkg/config/commandprovider"
	"cto-github.cisco.com/livdu/jupiter/pkg/config/consulprovider"
	"cto-github.cisco.com/livdu/jupiter/pkg/config/fileprovider"
	"cto-github.cisco.com/livdu/jupiter/pkg/consul"
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

//precedence from high to low
const (
	ConsulAppPrecedence            = iota
	ConsulDefaultPrecedence        = iota
	CommandlinePrecedence          = iota
	ApplicationLocalFilePrecedence = iota
	BootstrapLocalFilePrecedence   = iota
)

var applicationContext = NewContext()

var DefaultModule = &Module{
	Precedence: HighestPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(newConsulClient), fx.Provide(newBootstrapConfig), fx.Supply(applicationContext),
		fx.Invoke(bootstrap),
	},
}

func init() {
	Register(DefaultModule)
}

type bootstrapConfigParam struct {
	fx.In

	Config *config.Config `name:"bootstrap"`
}

type bootstrapConfigResult struct {
	fx.Out

	Config *config.Config `name:"bootstrap"`
}

func newConsulClient(p bootstrapConfigParam) *consul.Connection {
	connectionConfig := &consul.ConnectionConfig{}

	//TODO: error checking
	p.Config.Populate(connectionConfig, consul.ConfigRootConsulConnection)

	connection, _ := consul.NewConnection(connectionConfig)
	return connection
}

func newBootstrapConfig(cmd *cobra.Command) bootstrapConfigResult {
	//The order in which the providers are added determine their precedence
	fmt.Println("getting local configurations")
	commandLineProvider := commandprovider.NewCobraProvider("command line", CommandlinePrecedence, cmd, "cli.flag.")
	applicationContext.applicationConfig.AddProvider(commandLineProvider)

	//TODO: file provider can have different location, ext, and profile
	//TODO: error handling - if file doesn't exist
	fileProvider := fileprovider.NewFileProvidersFromBaseName("bootstrap file properties", BootstrapLocalFilePrecedence, "bootstrap", "yml")
	applicationContext.applicationConfig.AddProvider(fileProvider)

	applicationContext.applicationConfig.Load(false)

	return bootstrapConfigResult{Config:applicationContext.applicationConfig}
}

func bootstrap(lc fx.Lifecycle, consulConnection *consul.Connection) {
	fmt.Println("[bootstrap] - bootstrap")

	fileProvider := fileprovider.NewFileProvidersFromBaseName("application file properties", ApplicationLocalFilePrecedence, "application", "yml")
	applicationContext.applicationConfig.AddProvider(fileProvider)

	consulProviderConfig := &consulprovider.ConfigProviderConfig{}
	var err = applicationContext.applicationConfig.Populate(consulProviderConfig, consulprovider.ConfigRootConsulConfigProvider)
   	if err == nil && consulProviderConfig.Enabled {
   		//1. default contexts
		defaultContextConsulProvider := consulprovider.NewConsulProvider(
			"consul provider - default context",
			ConsulDefaultPrecedence,
			fmt.Sprintf("%s/%s", consulProviderConfig.Prefix, consulProviderConfig.DefaultContext),
			consulConnection,
			)
		applicationContext.applicationConfig.AddProvider(defaultContextConsulProvider)
   		//2. application specific contexts
		if appName, err := applicationContext.applicationConfig.Value(consulprovider.ConfigKeyAppName); err == nil {
			applicationContextConsulProvider := consulprovider.NewConsulProvider(
				"consul provider - app specific context",
				ConsulAppPrecedence,
				fmt.Sprintf("%s/%s", consulProviderConfig.Prefix, appName),
				consulConnection,
			)
			applicationContext.applicationConfig.AddProvider(applicationContextConsulProvider)
		}
   		//3. TODO: profile specific contexts

    }

	applicationContext.applicationConfig.Load(false)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ac := ctx.(*ApplicationContext)

			fmt.Println("[bootstrap] - On Application Start")
			ac.dumpConfigurations()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
