package init

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/commandprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/consulprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/envprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/fileprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/profile"
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var ConfigModule = &bootstrap.Module{
	Precedence: bootstrap.HighestPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(
			newCommandProvider,
			newOsEnvProvider,
			newBootstrapFileProvider,
			newBootstrapConfig,
			newApplicationFileProvider,
			newConsulProvider,
			newConsulConfigProperties,
			newApplicationConfig),
	},
}

func init() {
	bootstrap.Register(ConfigModule)
}

const (
	//preserve gap between different property sources to allow space for profile specific properties.
	precendencGap = 100

	//lower integer means higher precedence, therefore the list here is high to low in terms of precedence
	consulPrecedence               = iota * precendencGap
	commandlinePrecedence          = iota * precendencGap
	osEnvPrecedence				   = iota * precendencGap
	applicationLocalFilePrecedence = iota * precendencGap
	bootstrapLocalFilePrecedence   = iota * precendencGap
)

func newCommandProvider(cmd *cobra.Command) *commandprovider.ConfigProvider {
	p := commandprovider.NewCobraProvider(commandlinePrecedence, cmd, "cli.flag.")
	return p
}

func newOsEnvProvider() *envprovider.ConfigProvider {
	p := envprovider.NewEnvProvider(osEnvPrecedence)
	return p
}

type bootstrapFileProviderResult struct {
	fx.Out
	FileProvider []*fileprovider.ConfigProvider `name:"bootstrap_file_provider"`
}

func newBootstrapFileProvider() bootstrapFileProviderResult {
	name := "bootstrap"
	ext := "yml"

	precedence := bootstrapLocalFilePrecedence

	providers := make([]*fileprovider.ConfigProvider, 0, len(profile.Profiles) + 1)
	p := fileprovider.NewFileProvidersFromBaseName(precedence, name, ext)
	providers = append(providers, p)

	for _, profile := range profile.Profiles {
		precedence--
		p = fileprovider.NewFileProvidersFromBaseName(precedence, fmt.Sprintf("%s-%s", name, profile), ext)
		if p != nil {providers = append(providers, p)}
	}

	return bootstrapFileProviderResult{FileProvider: providers}
}

type bootstrapConfigParam struct {
	fx.In
	CmdProvider  *commandprovider.ConfigProvider
	OsEnvProvider *envprovider.ConfigProvider
	FileProvider []*fileprovider.ConfigProvider `name:"bootstrap_file_provider"`
}

func newBootstrapConfig(p bootstrapConfigParam) *appconfig.BootstrapConfig {
	providers := make([]appconfig.Provider, 0, len(p.FileProvider) + 1)

	providers = append(providers, p.CmdProvider, p.OsEnvProvider)

	for _, provider := range p.FileProvider {
		providers = append(providers, provider)
	}

	bootstrapConfig := appconfig.NewBootstrapConfig(providers...)

	error := bootstrapConfig.Load(false)
	if error != nil {
		panic(error)
	}

	return bootstrapConfig
}

func newConsulConfigProperties(bootstrapConfig *appconfig.BootstrapConfig) *consulprovider.ConsulConfigProperties {
	p := &consulprovider.ConsulConfigProperties{
		Prefix: "userviceconfiguration",
		DefaultContext: "defaultapplication",
		ProfileSeparator: ",",
	}
	bootstrapConfig.Bind(p, consulprovider.ConfigRootConsulConfigProvider)
	return p
}

func newConsulProvider(	bootstrapConfig *appconfig.BootstrapConfig, consulConfigProperties *consulprovider.ConsulConfigProperties, consulConnection *consul.Connection) []*consulprovider.ConfigProvider {
	appName := bootstrapConfig.Value(consulprovider.ConfigKeyAppName)

	providers := make([]*consulprovider.ConfigProvider, 0, len(profile.Profiles)*2 + 2)
	precedence := consulPrecedence

	//1. default contexts
	defaultContextConsulProvider := consulprovider.NewConsulProvider(
		precedence,
		fmt.Sprintf("%s/%s", consulConfigProperties.Prefix, consulConfigProperties.DefaultContext),
		consulConnection,
	)

	providers = append(providers, defaultContextConsulProvider)

	for _, profile := range profile.Profiles {
		precedence--
		p := consulprovider.NewConsulProvider(
			precedence,
			fmt.Sprintf("%s/%s%s%s",
				consulConfigProperties.Prefix, consulConfigProperties.DefaultContext, consulConfigProperties.ProfileSeparator, profile),
			consulConnection,
		)
		providers = append(providers, p)
	}

	precedence--

	//profile specific default context
	applicationContextConsulProvider := consulprovider.NewConsulProvider(
		precedence,
		fmt.Sprintf("%s/%s", consulConfigProperties.Prefix, appName),
		consulConnection,
	)

	for _, profile := range profile.Profiles {
		precedence--
		p := consulprovider.NewConsulProvider(
			precedence,
			fmt.Sprintf("%s/%s%s%s",
				consulConfigProperties.Prefix, appName, consulConfigProperties.ProfileSeparator, profile),
			consulConnection,
		)
		providers = append(providers, p)
	}

	providers = append(providers, applicationContextConsulProvider)

	return providers
}

type applicationFileProviderResult struct {
	fx.Out
	FileProvider []*fileprovider.ConfigProvider `name:"application_file_provider"`
}

func newApplicationFileProvider() applicationFileProviderResult {
	name := "application"
	ext := "yml"
	providers := make([]*fileprovider.ConfigProvider, 0, len(profile.Profiles) + 1)
	precedence := applicationLocalFilePrecedence
	p := fileprovider.NewFileProvidersFromBaseName(precedence, name, ext)
	providers = append(providers, p)

	for _, profile := range profile.Profiles {
		precedence--
		provider := fileprovider.NewFileProvidersFromBaseName(precedence, fmt.Sprintf("%s-%s", name, profile), ext)
		if provider != nil {
			providers = append(providers, provider)
		}
	}
	return applicationFileProviderResult{FileProvider: providers}
}

type newApplicationConfigParam struct {
	fx.In
	FileProvider       []*fileprovider.ConfigProvider `name:"application_file_provider"`
	ConsulProviders	   []*consulprovider.ConfigProvider
	BootstrapConfig    *appconfig.BootstrapConfig
}

func newApplicationConfig(p newApplicationConfigParam) *appconfig.ApplicationConfig {
	var mergedProvider []appconfig.Provider

	for _, provider := range p.FileProvider {
		mergedProvider = append(mergedProvider, provider)
	}
	for _, provider := range p.ConsulProviders {
		mergedProvider = append(mergedProvider, provider)
	}
	mergedProvider = append(mergedProvider, p.BootstrapConfig.Providers...)

	applicationConfig := appconfig.NewApplicationConfig(mergedProvider...)

	error := applicationConfig.Load(false)

	if error != nil {
		panic(error)
	}

	return applicationConfig
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}