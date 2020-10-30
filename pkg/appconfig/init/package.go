package init

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/bootstrap"
	"cto-github.cisco.com/livdu/jupiter/pkg/appconfig"
	"cto-github.cisco.com/livdu/jupiter/pkg/appconfig/commandprovider"
	"cto-github.cisco.com/livdu/jupiter/pkg/appconfig/consulprovider"
	"cto-github.cisco.com/livdu/jupiter/pkg/appconfig/fileprovider"
	"cto-github.cisco.com/livdu/jupiter/pkg/consul"
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var ConfigModule = &bootstrap.Module{
	Precedence: bootstrap.HighestPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(newCommandProvider),
		fx.Provide(newBootstrapFileProvider),
		fx.Provide(newBootstrapConfig),
		fx.Provide(newApplicationFileProvider),
		fx.Provide(newConsulProvider),
		fx.Provide(newConsulConfigProperties),
		fx.Invoke(loadApplicationConfig),
	},
}

func init() {
	bootstrap.Register(ConfigModule)
}

const (
	ConsulAppPrecedence            = iota
	ConsulDefaultPrecedence        = iota
	CommandlinePrecedence          = iota
	ApplicationLocalFilePrecedence = iota
	BootstrapLocalFilePrecedence   = iota
)

//TODO: each provide does the load and then adds itself to the config in application context.
// The invoke can be empty and just act as a trigger point
// This way the application context is growing as providers becomes ready
// The invoke can just add a flag to indicate that its fully loaded.

func newCommandProvider(cmd *cobra.Command) *commandprovider.ConfigProvider {
	p := commandprovider.NewCobraProvider("command line", CommandlinePrecedence, cmd, "cli.flag.")
	return p
}

type bootstrapFileProviderResult struct {
	fx.Out
	FileProvider *fileprovider.ConfigProvider `name:"bootstrap_file_provider"`
}

func newBootstrapFileProvider() bootstrapFileProviderResult {
	p := fileprovider.NewFileProvidersFromBaseName("bootstrap file properties", BootstrapLocalFilePrecedence, "bootstrap", "yml")
	return bootstrapFileProviderResult{FileProvider: p}
}

type bootstrapConfigParam struct {
	fx.In
	CmdProvider  *commandprovider.ConfigProvider
	FileProvider *fileprovider.ConfigProvider `name:"bootstrap_file_provider"`
}

type bootstrapConfigResult struct {
	fx.Out
	Config *appconfig.Config `name:"bootstrap_config"`
}

func newBootstrapConfig(p bootstrapConfigParam) bootstrapConfigResult {
	bootstrapConfig := appconfig.NewConfig(p.FileProvider, p.CmdProvider)
	bootstrapConfig.Load(false)

	return bootstrapConfigResult{Config: bootstrapConfig}
}

type consulConfigPropertiesParam struct {
	fx.In
	Config *appconfig.Config `name:"bootstrap_config"`
}

func newConsulConfigProperties(param consulConfigPropertiesParam) *consulprovider.ConsulConfigProperties {
	p := &consulprovider.ConsulConfigProperties{
		Prefix: "userviceconfiguration",
		DefaultContext: "defaultapplication",
	}
	param.Config.Bind(p, consulprovider.ConfigRootConsulConfigProvider)
	return p
}

type consulProviderParam struct {
	fx.In
	Config *appconfig.Config `name:"bootstrap_config"`
	ConsulConfigProperties *consulprovider.ConsulConfigProperties
	ConsulConnection *consul.Connection
}

type consulProviderResults struct {
	fx.Out
	Providers []appconfig.Provider `name:consul_providers`
}

func newConsulProvider(param consulProviderParam) consulProviderResults {
	consulConfigProperties := param.ConsulConfigProperties
	consulConnection := param.ConsulConnection
	appName, _ := param.Config.Value(consulprovider.ConfigKeyAppName)

	//1. default contexts
	defaultContextConsulProvider := consulprovider.NewConsulProvider(
		"consul provider - default context",
		ConsulDefaultPrecedence,
		fmt.Sprintf("%s/%s", consulConfigProperties.Prefix, consulConfigProperties.DefaultContext),
		consulConnection,
	)
	applicationContextConsulProvider := consulprovider.NewConsulProvider(
		"consul provider - app specific context",
		ConsulAppPrecedence,
		fmt.Sprintf("%s/%s", consulConfigProperties.Prefix, appName),
		consulConnection,
	)
	return consulProviderResults{Providers: []appconfig.Provider{defaultContextConsulProvider, applicationContextConsulProvider}}
}

type applicationFileProviderResult struct {
	fx.Out
	FileProvider *fileprovider.ConfigProvider `name:"application_file_provider"`
}

func newApplicationFileProvider() applicationFileProviderResult {
	p := fileprovider.NewFileProvidersFromBaseName("application file properties", ApplicationLocalFilePrecedence, "application", "yml")
	return applicationFileProviderResult{FileProvider: p}
}

type loadApplicationConfigParam struct {
	fx.In
	FileProvider       *fileprovider.ConfigProvider `name:"application_file_provider"`
	ConsulProviders	   []appconfig.Provider      `name:consul_providers`
	Config             *appconfig.Config            `name:"bootstrap_config"`
	ApplicationContext *bootstrap.ApplicationContext
}

func loadApplicationConfig(lc fx.Lifecycle, param loadApplicationConfigParam) {
	param.Config.AddProvider(param.FileProvider)
	for _, provider := range param.ConsulProviders {
		param.Config.AddProvider(provider)
	}
	param.Config.Load(false)
	param.ApplicationContext.UpdateConfig(param.Config)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}