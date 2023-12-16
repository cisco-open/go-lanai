// Package certsinit
// Initialize certificate manager with various of certificate sources
package certsinit

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	filecerts "cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs/source/file"
	"fmt"
	"go.uber.org/fx"
	"io"
)

const PropertiesPrefix = `tls`

var Module = &bootstrap.Module{
	Name:       "certs",
	Precedence: bootstrap.TlsConfigPrecedence,
	Options: []fx.Option{
		fx.Provide(BindProperties, ProvideDefaultManager),
		fx.Provide(
			filecerts.FxProvider(),
		),
		fx.Invoke(RegisterManagerLifecycle),
	},
}

func Use() {
	bootstrap.Register(Module)
}

type mgrDI struct {
	fx.In
	AppCfg    bootstrap.ApplicationConfig
	Props     certs.Properties
	Factories []certs.SourceFactory `group:"certs"`
}

func ProvideDefaultManager(di mgrDI) (certs.Manager, certs.Registrar) {
	reg := certs.NewDefaultManager(func(mgr *certs.DefaultManager) {
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
func BindProperties(appCfg bootstrap.ApplicationConfig) certs.Properties {
	props := certs.NewProperties()
	if e := appCfg.Bind(props, PropertiesPrefix); e != nil {
		panic(fmt.Errorf("failed to bind certificate properties: %v", e))
	}
	return *props
}

func RegisterManagerLifecycle(lc fx.Lifecycle, m certs.Manager) {
	lc.Append(fx.StopHook(func(context.Context) error {
		if closer, ok := m.(io.Closer); ok {
			return closer.Close()
		}
		return nil
	}))
}
