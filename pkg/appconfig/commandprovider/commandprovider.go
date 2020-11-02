package commandprovider

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/appconfig"
	"cto-github.cisco.com/livdu/jupiter/pkg/appconfig/args"
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

func (configProvider *ConfigProvider) Load() (loadError error) {
	defer func(){
		if loadError != nil {
			configProvider.IsLoaded = false
		} else {
			configProvider.IsLoaded = true
		}
	}()

	configProvider.once.Do(func() {
		configProvider.dynamicFlags = args.Extras(func(name string) bool {
			_, exists := configProvider.declaredFlags[name]
			return exists
		})
	})

	settings := make(map[string] interface{})

	for k, v := range configProvider.dynamicFlags {
		settings[k] = v
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


func NewCobraProvider(description string, precedence int, command *cobra.Command, prefix string) *ConfigProvider {

	flagSet := make(map[string]string)

	extractFlag := func(flag *pflag.Flag) {
		flagSet[prefix + flag.Name] = flag.Value.String()
	}

	command.InheritedFlags().VisitAll(extractFlag)
	command.LocalFlags().VisitAll(extractFlag)

	return &ConfigProvider{
		ProviderMeta:  appconfig.ProviderMeta{Description: description, Precedence: precedence},
		prefix:        prefix,
		declaredFlags: flagSet,
		appName:       command.Root().Name(),
	}
}