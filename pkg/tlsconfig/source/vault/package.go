package vaultcerts

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	certsource "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig/source"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"fmt"
	"go.uber.org/fx"
)

var logger = log.New("Certs.Vault")

const (
	sourceType = tlsconfig.SourceVault
)

type factoryDI struct {
	fx.In
	Props       tlsconfig.Properties
	VaultClient *vault.Client `optional:"true"`
}

func FxProvider() fx.Annotated {
	return fx.Annotated{
		Group: tlsconfig.FxGroup,
		Target: func(di factoryDI) (tlsconfig.SourceFactory, error) {
			if di.VaultClient == nil {
				logger.Warnf(`Vault Certificates source is not supported. Tips: Do not forget to initialize vault client.`)
				return nil, nil
			}

			rawDefaults, _ := di.Props.Sources[sourceType]
			factory, e := certsource.NewFactory[SourceProperties](sourceType, rawDefaults, func(props SourceProperties) tlsconfig.Source {
				// TODO review context
				return NewVaultProvider(context.Background(), di.VaultClient, props)
			})
			if e != nil {
				return nil, fmt.Errorf(`unable to register certificate source type [%s]: %v`, sourceType, e)
			}
			return factory, nil
		},
	}
}
