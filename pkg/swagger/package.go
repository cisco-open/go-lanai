package swagger

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"embed"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

//go:generate npm install --prefix nodejs
//go:generate npm run --prefix nodejs build --output_dir=../generated
//go:embed generated/*
var Content embed.FS

//go:embed defaults-swagger.yml
var defaultConfigFS embed.FS

var logger = log.New("Swagger")

var Module = &bootstrap.Module{
	Name: "swagger",
	Precedence: bootstrap.SwaggerPrecedence,
	PriorityOptions: []fx.Option{
		fx.Invoke(configureSecurity),
	},
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(bindSecurityProperties),
		fx.Provide(provideDiscoveryCustomizer),
		fx.Invoke(initialize),
	},
}

func Use() {
	bootstrap.Register(Module)
}

type initDI struct {
	fx.In
	Registrar *web.Registrar
	Properties SwaggerProperties
	Resolver   bootstrap.BuildInfoResolver `optional:"true"`
}

func initialize(di initDI) {
	di.Registrar.MustRegister(Content)
	di.Registrar.MustRegister(NewSwaggerController(di.Properties, di.Resolver))
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

type DiscoveryCustomizerDIOut struct {
	fx.Out

	Customizer discovery.Customizer `group:"discovery_customizer"`
}

func provideDiscoveryCustomizer() DiscoveryCustomizerDIOut {
	return DiscoveryCustomizerDIOut{
		Customizer: swaggerInfoDiscoveryCustomizer{},
	}
}