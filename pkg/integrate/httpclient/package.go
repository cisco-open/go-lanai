package httpclient

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("HttpClient")

var Module = &bootstrap.Module{
	Name: "swagger",
	Precedence: bootstrap.SwaggerPrecedence,
	Options: []fx.Option{
		//fx.Provide(bindSecurityProperties),
		fx.Provide(NewTestHttpClient),
		//fx.Invoke(Test),
	},
}

func Use() {
	bootstrap.Register(Module)
}