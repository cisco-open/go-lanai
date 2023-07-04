package opa

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	opatestserver "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test/server"
	"embed"
	"github.com/open-policy-agent/opa/plugins/bundle"
	oparest "github.com/open-policy-agent/opa/plugins/rest"
	sdktest "github.com/open-policy-agent/opa/sdk/test"
	"go.uber.org/fx"
	"io/fs"
)

/*************************
	Common Test Setup
 *************************/

//go:embed test/bundle/roles test/bundle/operations test/bundle/tenancy test/bundle/ownership test/bundle/api test/bundle/poc
var TestBundleFS embed.FS

const (
	bundleName       = "test-bundle"
	bundleServerName = "test-bundle-server"
)

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
		bundleFSs = []fs.FS{TestBundleFS}
	}
	return func(di BundleServerDI) (BundleServerOut, error) {
		server, e := opatestserver.NewBundleServer(di.AppCtx,
			opatestserver.WithBundleSources(bundleFSs...), opatestserver.WithBundleName(bundleName))
		if e != nil {
			return BundleServerOut{}, e
		}
		return BundleServerOut{
			Server:     server,
			Customizer: newConfigCustomizer(server, bundleName),
		}, nil
	}
}

type configCustomizer struct {
	Server     *sdktest.Server
	BundleName string
}

func newConfigCustomizer(server *sdktest.Server, bundleName string) *configCustomizer {
	return &configCustomizer{
		Server:     server,
		BundleName: bundleName,
	}
}

func (c configCustomizer) Customize(_ context.Context, cfg *Config) {
	cfg.Services = map[string]*oparest.Config{
		bundleServerName: {
			Name:             bundleServerName,
			URL:              c.Server.URL(),
			AllowInsecureTLS: true,
		},
	}
	cfg.Bundles = map[string]*bundle.Source{
		c.BundleName: {
			Service:  bundleServerName,
			Resource: "/bundles/" + c.BundleName,
		},
	}
}
