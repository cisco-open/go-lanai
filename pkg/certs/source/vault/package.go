package vaultcerts

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	certsource "cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs/source"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"encoding/json"
	"fmt"
	"go.uber.org/fx"
)

var logger = log.New("Certs.Vault")

const (
	sourceType = certs.SourceVault
)

var Module = &bootstrap.Module{
	Name:       "certs-vault",
	Precedence: bootstrap.TlsConfigPrecedence,
	Options: []fx.Option{
		fx.Provide(FxProvider()),
	},
}

func Use() {
	bootstrap.Register(Module)
}

type factoryDI struct {
	fx.In
	AppCtx      *bootstrap.ApplicationContext
	Props       certs.Properties `optional:"true"`
	VaultClient *vault.Client    `optional:"true"`
}

func FxProvider() fx.Annotated {
	return fx.Annotated{
		Group: certs.FxGroup,
		Target: func(di factoryDI) (certs.SourceFactory, error) {
			if di.VaultClient == nil {
				logger.Warnf(`Vault Certificates source is not supported. Tips: Do not forget to initialize vault client.`)
				return nil, nil
			}

			var rawDefaults json.RawMessage
			if di.Props.Sources != nil {
				rawDefaults, _ = di.Props.Sources[sourceType]
			}
			factory, e := certsource.NewFactory[SourceProperties](sourceType, rawDefaults, func(props SourceProperties) certs.Source {
				return NewVaultProvider(di.AppCtx, di.VaultClient, props)
			})
			if e != nil {
				return nil, fmt.Errorf(`unable to register certificate source type [%s]: %v`, sourceType, e)
			}
			return factory, nil
		},
	}
}
