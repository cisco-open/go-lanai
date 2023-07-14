package opatest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	opainit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/init"
	opatestserver "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test/server"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"embed"
	"github.com/open-policy-agent/opa/plugins/bundle"
	oparest "github.com/open-policy-agent/opa/plugins/rest"
	sdktest "github.com/open-policy-agent/opa/sdk/test"
	"go.uber.org/fx"
	"io/fs"
)

//go:embed bundle/**
var DefaultBundleFS embed.FS

const (
	TestBundleName = `test-bundle`
	TestBundlePathPrefix = `/bundles/`
	BundleServiceKey     = "test-bundle-service"
)

// WithBundles is a test.Options that initialize OPA and start a bundle server in test with given bundle FS.
// All FSs are built into a single bundle and loaded to OPA engine.
// If no bundle FS provided, DefaultBundleFS is used.
func WithBundles(bundleFSs ...fs.FS) test.Options {
	return test.WithOptions(
		apptest.WithModules(opainit.Module),
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
		server, e := opatestserver.NewBundleServer(di.AppCtx,
			opatestserver.WithBundleSources(bundleFSs...), opatestserver.WithBundleName(TestBundleName))
		if e != nil {
			return BundleServerOut{}, e
		}
		return BundleServerOut{
			Server:     server,
			Customizer: newConfigCustomizer(server, TestBundleName),
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

func (c configCustomizer) Customize(_ context.Context, cfg *opa.Config) {
	cfg.Services = map[string]*oparest.Config{
		BundleServiceKey: {
			Name:             BundleServiceKey,
			URL:              c.Server.URL(),
			AllowInsecureTLS: true,
		},
	}
	cfg.Bundles = map[string]*bundle.Source{
		c.BundleName: {
			Service:  BundleServiceKey,
			Resource: TestBundlePathPrefix + c.BundleName,
		},
	}
}
