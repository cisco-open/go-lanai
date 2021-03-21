package appconfig

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/commandprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/consulprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/envprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/fileprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/vaultprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/profile"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var logger = log.New("appconfig")

var ConfigModule = &bootstrap.Module{
	Name: "bootstrap endpoint",
	Precedence: bootstrap.AppConfigPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(
			newCommandProvider,
			newOsEnvProvider,
			newBootstrapFileProvider,
			newBootstrapConfig,
			newApplicationFileProvider,
			newConsulProvider,
			newConsulConfigProperties,
			newVaultProvider,
			newVaultConfigProperties,
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
	if p != nil {
		providers = append(providers, p)
	}

	for _, profile := range profile.Profiles {
		precedence--
		p = fileprovider.NewFileProvidersFromBaseName(precedence, fmt.Sprintf("%s-%s", name, profile), ext)
		if p != nil {providers = append(providers, p)}
	}

	if len(providers) == 0 {
		logger.Warnf("no bootstrap configuration file found. are you running from the project root directory?")
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
		Enabled: true,
	}
	bootstrapConfig.Bind(p, consulprovider.ConfigKeyConsulEndpoint)
	return p
}

type consulDi struct {
	fx.In
	BootstrapConfig *appconfig.BootstrapConfig
	ConsulConfigProperties *consulprovider.ConsulConfigProperties
	ConsulConnection *consul.Connection `optional:"true"`
}

func newConsulProvider(di consulDi) []*consulprovider.ConfigProvider {
	providers := make([]*consulprovider.ConfigProvider, 0, len(profile.Profiles)*2 + 2)
	if !di.ConsulConfigProperties.Enabled || di.ConsulConnection == nil {
		return providers
	}

	appName := di.BootstrapConfig.Value(bootstrap.PropertyKeyApplicationName)

	precedence := consulPrecedence

	//default contexts
	defaultContextConsulProvider := consulprovider.NewConsulProvider(
		precedence,
		fmt.Sprintf("%s/%s", di.ConsulConfigProperties.Prefix, di.ConsulConfigProperties.DefaultContext),
		di.ConsulConnection,
	)

	providers = append(providers, defaultContextConsulProvider)

	//profile specific default context
	for _, profile := range profile.Profiles {
		precedence--
		p := consulprovider.NewConsulProvider(
			precedence,
			fmt.Sprintf("%s/%s%s%s",
				di.ConsulConfigProperties.Prefix, di.ConsulConfigProperties.DefaultContext, di.ConsulConfigProperties.ProfileSeparator, profile),
			di.ConsulConnection,
		)
		providers = append(providers, p)
	}

	precedence--
	//app context
	applicationContextConsulProvider := consulprovider.NewConsulProvider(
		precedence,
		fmt.Sprintf("%s/%s", di.ConsulConfigProperties.Prefix, appName),
		di.ConsulConnection,
	)

	//profile specific app context
	for _, profile := range profile.Profiles {
		precedence--
		p := consulprovider.NewConsulProvider(
			precedence,
			fmt.Sprintf("%s/%s%s%s",
				di.ConsulConfigProperties.Prefix, appName, di.ConsulConfigProperties.ProfileSeparator, profile),
			di.ConsulConnection,
		)
		providers = append(providers, p)
	}

	providers = append(providers, applicationContextConsulProvider)

	return providers
}

func newVaultConfigProperties(bootstrapConfig *appconfig.BootstrapConfig) *vaultprovider.KvConfigProperties {
	p := &vaultprovider.KvConfigProperties{
		Backend: "secret",
		DefaultContext: "defaultapplication",
		ProfileSeparator: "/",
		Enabled: true,
		BackendVersion: 1,
	}
	bootstrapConfig.Bind(p, consulprovider.ConfigKeyConsulEndpoint)
	return p
}

type vaultDi struct {
	fx.In
	BootstrapConfig       *appconfig.BootstrapConfig
	VaultConfigProperties *vaultprovider.KvConfigProperties
	VaultClient           *vault.Client `optional:"true"`
}

func newVaultProvider(di vaultDi) []*vaultprovider.KeyValueConfigProvider {
	providers := make([]*vaultprovider.KeyValueConfigProvider, 0, len(profile.Profiles)*2 + 2)
	if !di.VaultConfigProperties.Enabled || di.VaultClient == nil{
		return providers
	}

	kvSecretEngine, err := vaultprovider.NewKvSecretEngine(
		di.VaultConfigProperties.BackendVersion, di.VaultConfigProperties.Backend, di.VaultClient)

	if err != nil {
		panic(err)
	}

	appName := di.BootstrapConfig.Value(bootstrap.PropertyKeyApplicationName)
	precedence := consulPrecedence

	//default contexts
	defaultContextConsulProvider := vaultprovider.NewVaultKvProvider(
		precedence,
		di.VaultConfigProperties.DefaultContext,
		kvSecretEngine,
	)

	providers = append(providers, defaultContextConsulProvider)

	for _, profile := range profile.Profiles {
		precedence--
		p := vaultprovider.NewVaultKvProvider(
			precedence,
			fmt.Sprintf("%s%s%s", di.VaultConfigProperties.DefaultContext, di.VaultConfigProperties.ProfileSeparator, profile),
			kvSecretEngine,
		)
		providers = append(providers, p)
	}

	precedence--

	//app context
	applicationContextConsulProvider := vaultprovider.NewVaultKvProvider(
		precedence,
		fmt.Sprintf("%s", appName),
		kvSecretEngine,
	)

	//profile specific app context
	for _, profile := range profile.Profiles {
		precedence--
		p := vaultprovider.NewVaultKvProvider(
			precedence,
			fmt.Sprintf("%s%s%s", appName, di.VaultConfigProperties.ProfileSeparator, profile),
			kvSecretEngine,
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
	if p != nil {
		providers = append(providers, p)
	}

	for _, profile := range profile.Profiles {
		precedence--
		provider := fileprovider.NewFileProvidersFromBaseName(precedence, fmt.Sprintf("%s-%s", name, profile), ext)
		if provider != nil {
			providers = append(providers, provider)
		}
	}

	if len(providers) == 0 {
		logger.Warnf("no application configuration file found. are you running from the project root directory?")
	}
	return applicationFileProviderResult{FileProvider: providers}
}

type newApplicationConfigParam struct {
	fx.In
	FileProvider       []*fileprovider.ConfigProvider `name:"application_file_provider"`
	ConsulProviders	   []*consulprovider.ConfigProvider
	VaultProviders     []*vaultprovider.KeyValueConfigProvider
	BootstrapConfig    *appconfig.BootstrapConfig
}

type appConfigDIOut struct {
	fx.Out
	ACPtr *appconfig.ApplicationConfig
	ACI   bootstrap.ApplicationConfig
}

// expose *appconfig.ApplicationConfig as both pointer and interface
func newApplicationConfig(p newApplicationConfigParam) appConfigDIOut {
	var mergedProvider []appconfig.Provider

	for _, provider := range p.FileProvider {
		mergedProvider = append(mergedProvider, provider)
	}
	for _, provider := range p.ConsulProviders {
		mergedProvider = append(mergedProvider, provider)
	}
	for _, provider := range p.VaultProviders {
		mergedProvider = append(mergedProvider, provider)
	}
	mergedProvider = append(mergedProvider, p.BootstrapConfig.Providers()...)

	applicationConfig := appconfig.NewApplicationConfig(mergedProvider...)

	error := applicationConfig.Load(false)

	if error != nil {
		panic(error)
	}

	return appConfigDIOut{
		ACPtr: applicationConfig,
		ACI: applicationConfig,
	}
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}