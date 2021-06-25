package passwdidp

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"embed"
	"go.uber.org/fx"
)

//var logger = log.New("SEC.Passwd")

const (
	OrderWhiteLabelTemplateFS = 0
	OrderTemplateFSOverwrite  = OrderWhiteLabelTemplateFS - 1000
)

//go:embed web/whitelabel/*
var whiteLabelContent embed.FS

//go:embed defaults-passwd-auth.yml
var defaultConfigFS embed.FS

//goland:noinspection GoNameStartsWithPackageName
var PasswdIdpModule = &bootstrap.Module{
	Name: "error handling",
	Precedence: security.MaxSecurityPrecedence - 100,
	Options: []fx.Option {
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(BindPwdAuthProperties),
		fx.Invoke(register),
	},
}

func Use() {
	bootstrap.Register(PasswdIdpModule)
}

func register(r *web.Registrar) {
	r.MustRegister(web.OrderedFS(whiteLabelContent, OrderWhiteLabelTemplateFS))
}
