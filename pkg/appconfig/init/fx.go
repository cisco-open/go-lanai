package appconfig

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/fileprovider"
	"embed"
	"go.uber.org/fx"
	"path/filepath"
)

const (
	FxGroupBootstrap = "bootstrap-config"
	FxGroupApplication = "application-config"
	FxGroupDefaults = "default-config"
)

func FxEmbeddedDefaults(fs embed.FS) fx.Option {
	return embeddedFileProviderFxOptions(FxGroupDefaults, fs)
}

func FxEmbeddedApplicationAdHoc(fs embed.FS) fx.Option {
	return embeddedFileProviderFxOptions(FxGroupApplication, fs)
}

func FxEmbeddedBootstrapAdHoc(fs embed.FS) fx.Option {
	return embeddedFileProviderFxOptions(FxGroupBootstrap, fs)
}

func embeddedFileProviderFxOptions(fxGroup string, fs embed.FS) fx.Option {
	files, e := fs.ReadDir(".")
	if e != nil {
		return fx.Supply()
	}

	const ext = "yml"
	providers := make([]interface{}, 0)
	for _, f := range files {
		if !f.IsDir() || filepath.Ext(f.Name()) == ext {
			providers = append(providers, fxEmbeddedFileProvider(fxGroup, f.Name(), fs))
		}
	}
	return fx.Provide(providers...)
}

func fxEmbeddedFileProvider(fxGroup string, filepath string, fs embed.FS) fx.Annotated {
	fn := func() appconfig.Provider{
		// Note order will be overwritten by corresponding provider group
		provider, _ := fileprovider.NewEmbeddedFSProvider(0, filepath, fs)
		return provider
	}

	return fx.Annotated {
		Group: fxGroup,
		Target: fn,
	}
}
