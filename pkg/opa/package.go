package opa

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"embed"
	"go.uber.org/fx"
)

var logger = log.New("OPA")

//go:embed defaults-opa.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Precedence: bootstrap.SecurityPrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(BindProperties, ProvideEmbeddedOPA),
		fx.Invoke(InitializeEmbeddedOPA),
	},
}

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}


