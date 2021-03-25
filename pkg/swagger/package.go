package swagger

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
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
	PriorityOptions: []fx.Option{
		fx.Invoke(configureSecurity),
	},
	Options: []fx.Option{
		fx.Provide(bindSecurityProperties),
		fx.Invoke(initialize),
	},
}

func Use() {
	bootstrap.Register(Module)
}

func initialize(registrar *web.Registrar, prop SwaggerProperties) {
	registrar.Register(Content)
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

type secDI struct {
	fx.In
	SecRegistrar security.Registrar `optional:"true"`
}

// configureSecurity register security.Configurer that control how security works on endpoints
func configureSecurity(di secDI) {
	if di.SecRegistrar != nil {
		di.SecRegistrar.Register(&swaggerSecurityConfigurer{})
	}
}