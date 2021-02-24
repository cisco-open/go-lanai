package commandprovider

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/args"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sync"
)

const (
	configKeyAppName = "spring.application.name"
)

type ConfigProvider struct {
	appconfig.ProviderMeta
	prefix        string
	appName       string
	declaredFlags map[string]string //flags pre-declared by our command (e.g. --help). Cobra will parse these.
	dynamicFlags  map[string]string //flags not declared by us, so we need to parse these ourselves.
	once          sync.Once
}

func (configProvider *ConfigProvider) Name() string {
	return "command-line"
}

func (configProvider *ConfigProvider) Load() (loadError error) {
	defer func(){
		if loadError != nil {
			configProvider.Loaded = false
		} else {
			configProvider.Loaded = true
		}
	}()

	configProvider.once.Do(func() {
		configProvider.dynamicFlags = args.Extras(func(name string) bool {
			_, exists := configProvider.declaredFlags[name]
			return exists
		})
	})

	settings := make(map[string] interface{})

	for k, v := range configProvider.declaredFlags {
		settings[configProvider.prefix + k] = v
	}

	for k, v := range configProvider.dynamicFlags {
		settings[k] = v
	}

	// Apply application name
	settings[configKeyAppName] = configProvider.appName

	unFlattened, loadError := appconfig.UnFlatten(settings)

	if loadError != nil {
		return loadError
	}

	configProvider.Settings = unFlattened

	return nil
}


func NewCobraProvider(precedence int, command *cobra.Command, prefix string) *ConfigProvider {

	flagSet := make(map[string]string)

	extractFlag := func(flag *pflag.Flag) {
		flagSet[flag.Name] = flag.Value.String()
	}

	command.InheritedFlags().VisitAll(extractFlag)
	command.LocalFlags().VisitAll(extractFlag)

	return &ConfigProvider{
		ProviderMeta:  appconfig.ProviderMeta{Precedence: precedence},
		prefix:        prefix,
		declaredFlags: flagSet,
		appName:       command.Root().Name(),
	}
}