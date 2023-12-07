// Package tlsconfiginit
// Initialize certificate manager with various of certificate sources
package tlsconfiginit

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	acmcerts "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig/source/acm"
	filecerts "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig/source/file"
	vaultcerts "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig/source/vault"
	"fmt"
	"go.uber.org/fx"
)

const PropertiesPrefix = `tls`

var Module = &bootstrap.Module{
	Name:       "tls-config",
	Precedence: bootstrap.TlsConfigPrecedence,
	Options: []fx.Option{
		fx.Provide(BindProperties, ProvideDefaultManager),
		// TODO maybe we don't automatically register all sources
		fx.Provide(
			filecerts.FxProvider(),
			vaultcerts.FxProvider(),
			acmcerts.FxProvider(),
		),
	},
}

func Use() {
	bootstrap.Register(Module)
}

type mgrDI struct {
	fx.In
	AppCfg    bootstrap.ApplicationConfig
	Props     tlsconfig.Properties
	Factories []tlsconfig.SourceFactory `group:"certs"`
}

func ProvideDefaultManager(di mgrDI) (tlsconfig.Manager, tlsconfig.Registrar) {
	reg := tlsconfig.NewDefaultManager(func(mgr *tlsconfig.DefaultManager) {
		mgr.ConfigLoaderFunc = di.AppCfg.Bind
		mgr.Properties = di.Props
	})
	for _, f := range di.Factories {
		if f != nil {
			reg.MustRegister(f)
		}
	}
	return reg, reg
}

// BindProperties create and bind SessionProperties, with a optional prefix
func BindProperties(appCfg bootstrap.ApplicationConfig) tlsconfig.Properties {
	var props tlsconfig.Properties
	if e := appCfg.Bind(&props, PropertiesPrefix); e != nil {
		panic(fmt.Errorf("failed to bind certificate properties: %v", e))
	}
	return props
}
