package cliprovider

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
)

const (
	defaultConfigSearchPath = "configs"
)

type StaticConfigProvider struct {
	appconfig.ProviderMeta
	appName string
}

func NewStaticConfigProvider(order int, execCtx *bootstrap.CliExecContext) *StaticConfigProvider {
	return &StaticConfigProvider{
		ProviderMeta: appconfig.ProviderMeta{
			Precedence: order,
		},
		appName: execCtx.Cmd.Root().Name(),
	}
}

func (p *StaticConfigProvider) Name() string {
	return "default"
}

func (p *StaticConfigProvider) Load() (err error) {
	defer func(){
		p.Loaded = err == nil
	}()

	settings := map[string]interface{}{}

	// Apply application name, profiles, etc
	settings[appconfig.PropertyKeyApplicationName] = p.appName
	settings[appconfig.PropertyKeyConfigFileSearchPath] = []string{defaultConfigSearchPath}

	// un-flatten
	unFlattened, err := appconfig.UnFlatten(settings)
	if err == nil {
		p.Settings = unFlattened
	}
	return
}


