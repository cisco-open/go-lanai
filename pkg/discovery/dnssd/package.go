package dnssd

import (
	"context"
	"embed"
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/log"
	"go.uber.org/fx"
	"io"
)

var logger = log.New("SD.DNS")

//go:embed defaults-discovery.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name:       "consul service discovery",
	Precedence: bootstrap.ServiceDiscoveryPrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(
			BindDiscoveryProperties,
			provideDiscoveryClient),
		fx.Invoke(closeDiscoveryClient),
	},
}

func Use() {
	bootstrap.Register(Module)
}

func provideDiscoveryClient(ctx *bootstrap.ApplicationContext, props DiscoveryProperties) discovery.Client {
	return NewDiscoveryClient(ctx, func(opt *ClientConfig) {
		opt.DNSServerAddr = props.Addr
		opt.SRVTargetTemplate = props.SRVTargetTemplate
		opt.SRVProto = props.SRVProto
		opt.SRVService = props.SRVService
	})
}

func closeDiscoveryClient(lc fx.Lifecycle, client discovery.Client) {
	lc.Append(fx.StopHook(func(ctx context.Context) error {
		return client.(io.Closer).Close()
	}))
}
