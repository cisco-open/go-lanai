package envprovider

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"os"
	"strings"
)

type ConfigProvider struct {
	appconfig.ProviderMeta
}

const dot = rune('.')

func (configProvider *ConfigProvider) Name() string {
	return "environment-variable"
}

func (configProvider *ConfigProvider) Load(_ context.Context) (loadError error) {
	defer func() {
		if loadError != nil {
			configProvider.Loaded = false
		} else {
			configProvider.Loaded = true
		}
	}()

	flatSettings := make(map[string]interface{})

	for _, e := range os.Environ() {
		kv := strings.SplitN(e, "=", 2)
		k := kv[0]
		v := kv[1]

		var runes []rune
		for pos, char := range k {
			if strings.Compare(string(char), "_") == 0 {
				if pos>0 && strings.Compare(string(runes[pos-1]) , "_") != 0 {
					runes = append(runes, dot)
				} else {
					runes = append(runes, char)
				}
			} else {
				runes = append(runes, char)
			}
		}

		flatSettings[string(runes)] = utils.ParseString(v)
	}

	unFlattenedSettings, loadError := appconfig.UnFlatten(flatSettings)
	if loadError != nil {
		return loadError
	}

	configProvider.Settings = unFlattenedSettings
	return nil
}

func NewEnvProvider(precedence int) *ConfigProvider {
	return &ConfigProvider{
		ProviderMeta: appconfig.ProviderMeta{Precedence: precedence},
	}
}
