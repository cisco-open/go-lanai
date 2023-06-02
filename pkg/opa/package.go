package opa

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"embed"
	"go.uber.org/fx"
)

var logger = log.New("OPA")

//go:embed opa-config.yml
var ConfigFS embed.FS

var Module = &bootstrap.Module{
	Precedence: bootstrap.SecurityPrecedence,
	Options: []fx.Option{
		fx.Provide(ProvideEmbeddedOPA),
		fx.Provide(ProvideBundleServer),
		fx.Invoke(InitializeEmbeddedOPA),
		fx.Invoke(InitializeBundleServer),
	},
}

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}


