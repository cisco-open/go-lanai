package appconfig

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/fileprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"fmt"
)

func newApplicationFileProviderGroup() appConfigProvidersOut {
	const name = "application"
	const ext = "yml"
	group := appconfig.NewProfileBasedProviderGroup(applicationLocalFilePrecedence)
	group.KeyFunc = func(profile string) string {
		if profile == "" {
			return name
		}
		return fmt.Sprintf("%s-%s", name, profile)
	}
	group.CreateFunc = func(name string, order int, conf bootstrap.ApplicationConfig) appconfig.Provider {
		ptr, exists := fileprovider.NewFileProvidersFromBaseName(order, name, ext, conf)
		if !exists || ptr == nil {
			return nil
		}
		return ptr
	}
	group.ProcessFunc = func(ctx context.Context, providers []appconfig.Provider) []appconfig.Provider {
		if len(providers) == 0 {
			logger.Warnf("no application configuration file found. are you running from the project root directory?")
		}
		return providers
	}

	return appConfigProvidersOut {
		ProviderGroup: group,
	}
}
