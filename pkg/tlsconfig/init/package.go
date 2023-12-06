// Package tlsconfiginit
// This is an internal package. Do not use outside of go-lanai
package tlsconfiginit

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	acmcerts "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig/source/acm"
	filecerts "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig/source/file"
	vaultcerts "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig/source/vault"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name:       "tls-config",
	Precedence: bootstrap.TlsConfigPrecedence,
	Options: []fx.Option{
		fx.Provide(tlsconfig.BindProperties, ProvideDefaultManager),
		// TODO maybe we don't automatically register all sources
		fx.Provide(
			filecerts.FxProvider(),
			vaultcerts.FxProvider(),
			acmcerts.FxProvider(),
		),
		// Remove me
		fx.Provide(NewProviderFactory),
	},
}

func Use() {
	bootstrap.Register(Module)
}

type mgrDI struct {
	fx.In
	AppCfg bootstrap.ApplicationConfig
	Factories []tlsconfig.SourceFactory `group:"certs"`
}

func ProvideDefaultManager(di mgrDI) (tlsconfig.Manager, tlsconfig.Registrar) {
	reg := tlsconfig.NewDefaultManager(di.AppCfg)
	for _, f := range di.Factories {
		if f != nil {
			reg.MustRegister(f)
		}
	}
	return reg, reg
}

type factoryDi struct {
	fx.In
	AppCtx *bootstrap.ApplicationContext
	Manager tlsconfig.Manager
}

// Deprecated
func NewProviderFactory(di factoryDi) *tlsconfig.ProviderFactory {
	return &tlsconfig.ProviderFactory{
		Manager: di.Manager,
		AppCtx:  di.AppCtx,
	}
}
