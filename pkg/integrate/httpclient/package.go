package httpclient

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"go.uber.org/fx"
	"time"
)

var logger = log.New("HttpClient")

var Module = &bootstrap.Module{
	Name: "http-client",
	Precedence: bootstrap.HttpClientPrecedence,
	Options: []fx.Option{
		fx.Provide(bindHttpClientProperties),
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
	Properties  HttpClientProperties
	DiscClient  discovery.Client
	Customizers []ClientCustomizer `group:"http-client"`
}

func provideHttpClient(di clientDI) Client {
	options := []ClientOptions{func(opt *ClientOption) {
		opt.MaxRetries = di.Properties.MaxRetries
		opt.Timeout = time.Duration(di.Properties.Timeout)
		opt.Logging.Level = di.Properties.Logger.Level
		opt.Logging.DetailsLevel = di.Properties.Logger.DetailsLevel
		opt.Logging.SanitizeHeaders = utils.NewStringSet(di.Properties.Logger.SanitizeHeaders...)
		opt.Logging.ExcludeHeaders = utils.NewStringSet(di.Properties.Logger.ExcludeHeaders...)
	}}
	for _, customizer := range di.Customizers {
		options = append(options, customizer.Customize)
	}

	return newClient(di.DiscClient, options...)
}