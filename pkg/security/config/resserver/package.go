package resserver

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/timeoutsupport"
	"embed"
	"go.uber.org/fx"
)

//go:embed defaults-resserver.yml
var defaultConfigFS embed.FS

//goland:noinspection GoNameStartsWithPackageName
var Module = &bootstrap.Module{
	Name: "oauth2 authserver",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(jwt.BindCryptoProperties),
		fx.Provide(ProvideResServerDI),
		fx.Invoke(ConfigureResourceServer),
	},
}

func Use() {
	security.Use()
	bootstrap.Register(Module)
	timeoutsupport.Use()
}

