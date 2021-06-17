package cliprovider

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/args"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"fmt"
	"github.com/spf13/pflag"
	"sync"
)

var (
	declaredFlagsMapping = map[string]string{
		bootstrap.CliFlagActiveProfile:     appconfig.PropertyKeyActiveProfiles,
		bootstrap.CliFlagAdditionalProfile: appconfig.PropertyKeyAdditionalProfiles,
		bootstrap.CliFlagConfigSearchPath:  appconfig.PropertyKeyConfigFileSearchPath,
	}
)

type ConfigProvider struct {
	appconfig.ProviderMeta
	prefix        string
	declaredFlags map[string]interface{} // flags pre-declared by our command (e.g. --help). Cobra will parse these.
	args          []string // cobra arguments (args after standalone "--" )
	dynamicFlags  map[string]string // flags not declared by us, so we need to parse these ourselves.
	kvArgs        map[string]string // key=value pairs from cobra arguments (args after standalone "--" )
	once          sync.Once
}

func (configProvider *ConfigProvider) Name() string {
	return "command-line"
}

func (configProvider *ConfigProvider) Load(_ context.Context) (loadError error) {
	defer func() {
		configProvider.Loaded = loadError == nil
	}()

	configProvider.once.Do(func() {
		configProvider.dynamicFlags = args.ExtraFlags(func(name string) bool {
			_, exists := configProvider.declaredFlags[name]
			return exists
		})
		configProvider.kvArgs = args.ExtraKVArgs(configProvider.args)
	})

	settings := make(map[string]interface{})

	// dynamic flags
	for k, v := range configProvider.dynamicFlags {
		settings[k] = v
	}

	// declared flags
	for k, v := range configProvider.declaredFlags {
		v = configProvider.convertDeclaredFlag(v)
		settings[configProvider.prefix+k] = v
		if pk, ok := declaredFlagsMapping[k]; ok {
			settings[pk] = v
		}
	}

	// arguments
	for k, v := range configProvider.kvArgs {
		settings[k] = v
	}

	// un-flatten
	unFlattened, loadError := appconfig.UnFlatten(settings)

	if loadError != nil {
		return loadError
	}

	configProvider.Settings = unFlattened

	return nil
}

func (configProvider *ConfigProvider) convertDeclaredFlag(value interface{}) interface{} {
	switch v := value.(type) {
	case pflag.SliceValue:
		return v.GetSlice()
	case pflag.Value:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func NewCobraProvider(precedence int, execCtx *bootstrap.CliExecContext, prefix string) *ConfigProvider {

	flagSet := make(map[string]interface{})

	extractFlag := func(flag *pflag.Flag) {
		if flag.Changed {
			flagSet[flag.Name] = flag.Value
		}
	}

	execCtx.Cmd.InheritedFlags().VisitAll(extractFlag)
	execCtx.Cmd.LocalFlags().VisitAll(extractFlag)

	return &ConfigProvider{
		ProviderMeta:  appconfig.ProviderMeta{Precedence: precedence},
		prefix:        prefix,
		declaredFlags: flagSet,
		args:          execCtx.Args,
	}
}
