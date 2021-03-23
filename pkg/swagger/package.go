package swagger

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"embed"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

//go:embed frontend/*
var Content embed.FS

var logger = log.New("Swagger")

var Module = &bootstrap.Module{
	Name: "swagger",
	Precedence: bootstrap.SwaggerPrecedence,
	Options: []fx.Option{
		fx.Provide(bindSecurityProperties),
		fx.Invoke(initialize),
	},
}

func Use() {}

func init() {
	bootstrap.Register(Module)
}

func initialize(registrar *web.Registrar, prop SwaggerProperties) {
	registrar.AddEmbeddedFs(Content)
	controller := NewSwaggerController(prop)
	registrar.Register(controller.Mappings())
}

func bindSecurityProperties(ctx *bootstrap.ApplicationContext) SwaggerProperties {
	props := NewSwaggerSsoProperties()
	if err := ctx.Config().Bind(props, SwaggerPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind SwaggerSsoProperties"))
	}
	return *props
}