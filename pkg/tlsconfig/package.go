// Package tlsconfig
// This is an internal package. Do not use outside of go-lanai
package tlsconfig

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"go.uber.org/fx"
)

var logger = log.New("tlsconfig")

var Module = &bootstrap.Module{
	Name: "tls-config",
	Precedence: bootstrap.TlsConfigPrecedence,
	Options: []fx.Option{
		fx.Provide(NewProviderFactory),
	},
}

type factoryDi struct {
	fx.In
	Vc *vault.Client `optional:"true"`
}

func NewProviderFactory(di factoryDi) *ProviderFactory {
	return &ProviderFactory{
		vc: di.Vc,
	}
}
