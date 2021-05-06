package passwdidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"embed"
	"go.uber.org/fx"
)

var logger = log.New("SEC.Passwd")

const (
	OrderWhiteLabelTemplateFS = 0
	OrderTemplateFSOverwrite  = OrderWhiteLabelTemplateFS - 1000
)

//goland:noinspection GoNameStartsWithPackageName
var PasswdIdpModule = &bootstrap.Module{
	Name: "error handling",
	Precedence: security.MaxSecurityPrecedence - 100,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

//go:embed web/whitelabel/*
var WhiteLabelContent embed.FS

func init() {
	bootstrap.Register(PasswdIdpModule)
}

func register(r *web.Registrar) {
	r.MustRegister(web.OrderedFS(WhiteLabelContent, OrderWhiteLabelTemplateFS))
}
