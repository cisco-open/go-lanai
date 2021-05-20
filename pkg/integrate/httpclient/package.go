package httpclient

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("HttpClient")

var Module = &bootstrap.Module{
	Name: "http-client",
	Precedence: bootstrap.HttpClientPrecedence,
	Options: []fx.Option{
		//fx.Provide(bindSecurityProperties),
		fx.Provide(provideHttpClient),
	},
}

func Use() {
	bootstrap.Register(Module)
}

// FxClientCustomizers takes providers of ClientCustomizer and wrap them with FxGroup
func FxClientCustomizers(providers ...interface{}) []fx.Annotated {
	annotated := make([]fx.Annotated, len(providers))
	for i, t := range providers {
		annotated[i] = fx.Annotated{
			Group:  FxGroup,
			Target: t,
		}
	}
	return annotated
}

type clientDI struct {
	fx.In
	DiscClient discovery.Client
	Customizers []ClientCustomizer `group:"http-client"`
}

func provideHttpClient(di clientDI) Client {
	// TODO use properties
	options := []ClientOptions{func(opt *ClientOption) {
		opt.Verbose = true
	}}
	for _, customizer := range di.Customizers {
		options = append(options, customizer.Customize)
	}

	return NewClient(di.DiscClient, options...)
}