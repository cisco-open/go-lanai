package opa

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	opatestserver "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test/server"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"embed"
	"fmt"
	"github.com/open-policy-agent/opa/plugins/bundle"
	oparest "github.com/open-policy-agent/opa/plugins/rest"
	sdktest "github.com/open-policy-agent/opa/sdk/test"
	"go.uber.org/fx"
	"io/fs"
)

/*************************
	Common Test Setup
 *************************/

//go:embed test/bundle/.manifest test/bundle/roles test/bundle/operations test/bundle/tenancy test/bundle/ownership test/bundle/api test/bundle/poc
var TestBundleFS embed.FS

type BundleServerDI struct {
	fx.In
	AppCtx *bootstrap.ApplicationContext
}

type BundleServerOut struct {
	fx.Out
	Server     *sdktest.Server
	Customizer ConfigCustomizer `group:"opa"`
}

func BundleServerProvider(bundleFSs ...fs.FS) func(BundleServerDI) (BundleServerOut, error) {
	if len(bundleFSs) == 0 {
		bundleFSs =  []fs.FS{TestBundleFS}
	}
	return func(di BundleServerDI) (BundleServerOut, error) {
		opts := make([]opatestserver.BundleServerOptions, 0, len(bundleFSs) + 1)
		names := make([]string, 0, len(bundleFSs) + 1)
		for _, fsys := range bundleFSs {
			name := fmt.Sprintf("/bundles/test-%s", utils.RandomString(6))
			opts = append(opts, opatestserver.WithBundleFS(name, fsys))
			names = append(names, name)
		}

		server, e := opatestserver.NewBundleServer(di.AppCtx, opts...)
		if e != nil {
			return BundleServerOut{}, e
		}
		return BundleServerOut{
			Server:     server,
			Customizer: newConfigCustomizer(server, names),
		}, nil
	}
}

type configCustomizer struct {
	Server      *sdktest.Server
	BundleNames []string
}

func newConfigCustomizer(server *sdktest.Server, bundleNames []string) *configCustomizer {
	return &configCustomizer{
		Server: server,
		BundleNames: bundleNames,
	}
}

func (c configCustomizer) Customize(_ context.Context, cfg *Config) {
	const serviceKey = `test-bundle-server`
	cfg.Services = map[string]*oparest.Config{
		serviceKey: {
			Name:             serviceKey,
			URL:              c.Server.URL(),
			AllowInsecureTLS: true,
		},
	}
	cfg.Bundles = map[string]*bundle.Source{}
	for _, name := range c.BundleNames {
		src := &bundle.Source{
			Service:        serviceKey,
			Resource:       name,
		}
		cfg.Bundles[name] = src
	}
}