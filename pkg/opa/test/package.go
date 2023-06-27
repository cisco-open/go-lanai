package opatest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	opatestserver "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test/server"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"embed"
	"fmt"
	"github.com/open-policy-agent/opa/plugins/bundle"
	oparest "github.com/open-policy-agent/opa/plugins/rest"
	sdktest "github.com/open-policy-agent/opa/sdk/test"
	"go.uber.org/fx"
	"io/fs"
)

//go:embed bundle/**
var DefaultBundleFS embed.FS

const (
	bundleServiceKey = "test-bundle-service"
)

// WithBundles is a test.Options that initialize OPA and start a bundle server in test with given bundle FS.
// Each FS are built into a single bundle and loaded to OPA engine.
// If no bundle FS provided, DefaultBundleFS is used.
func WithBundles(bundleFSs ...fs.FS) test.Options {
	return test.WithOptions(
		apptest.WithModules(opa.Module),
		apptest.WithFxOptions(
			fx.Provide(BundleServerProvider(bundleFSs...)),
			fx.Invoke(opatestserver.InitializeBundleServer),
		),
	)
}

type BundleServerDI struct {
	fx.In
	AppCtx *bootstrap.ApplicationContext
}

type BundleServerOut struct {
	fx.Out
	Server     *sdktest.Server
	Customizer opa.ConfigCustomizer `group:"opa"`
}

func BundleServerProvider(bundleFSs ...fs.FS) func(BundleServerDI) (BundleServerOut, error) {
	if len(bundleFSs) == 0 {
		bundleFSs = []fs.FS{DefaultBundleFS}
	}
	return func(di BundleServerDI) (BundleServerOut, error) {
		opts := make([]opatestserver.BundleServerOptions, 0, len(bundleFSs))
		names := make([]string, 0, len(bundleFSs))
		for _, fsys := range bundleFSs {
			name := fmt.Sprintf("/bundles/test-%s", utils.RandomString(6))
			if fsys == DefaultBundleFS {
				name = "/bundles/test-default"
			}
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
		Server:      server,
		BundleNames: bundleNames,
	}
}

func (c configCustomizer) Customize(_ context.Context, cfg *opa.Config) {
	cfg.Services = map[string]*oparest.Config{
		bundleServiceKey: {
			Name:             bundleServiceKey,
			URL:              c.Server.URL(),
			AllowInsecureTLS: true,
		},
	}
	cfg.Bundles = map[string]*bundle.Source{}
	for _, name := range c.BundleNames {
		src := &bundle.Source{
			Service:  bundleServiceKey,
			Resource: name,
		}
		cfg.Bundles[name] = src
	}
}
