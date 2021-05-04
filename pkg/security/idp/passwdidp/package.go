package passwdidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/assets"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
	"embed"
	"go.uber.org/fx"
)

var logger = log.New("SEC.Passwd")

//goland:noinspection GoNameStartsWithPackageName
var PasswdIdpModule = &bootstrap.Module{
	Name: "error handling",
	Precedence: security.MaxSecurityPrecedence - 100,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

//go:generate npm install --prefix web/nodejs
//go:generate go run github.com/mholt/archiver/cmd/arc -overwrite -folder-safe=false unarchive web/nodejs/node_modules/@msx/login-app/login-app-ui.zip web/login-ui/
//go:embed web/login-ui/*
var GeneratedContent embed.FS

//go:embed web/whitelabel/*
var WhiteLabelContent embed.FS

func init() {
	bootstrap.Register(PasswdIdpModule)
}

func register(r *web.Registrar) {
	r.MustRegister(web.OrderedFS(GeneratedContent, order.Highest + 1000))
	r.MustRegister(web.OrderedFS(WhiteLabelContent, order.Highest + 2000))
	r.MustRegister(assets.New("app", "web/login-ui"))
	r.MustRegister(NewLoginFormController())
	r.MustRegister(template.New().
			Get("/error").
			HandlerFunc(errorhandling.ErrorWithStatus).
			Build(),
	)
}
