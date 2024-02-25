// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package opatest

import (
    "context"
    "embed"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/opa"
    opainit "github.com/cisco-open/go-lanai/pkg/opa/init"
    opatestserver "github.com/cisco-open/go-lanai/pkg/opa/test/server"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/open-policy-agent/opa/plugins/bundle"
    oparest "github.com/open-policy-agent/opa/plugins/rest"
    sdktest "github.com/open-policy-agent/opa/sdk/test"
    "go.uber.org/fx"
    "io/fs"
)

//go:embed bundle/**
var DefaultBundleFS embed.FS

//go:embed test-defaults-opa.yml
var DefaultConfigFS embed.FS

const (
	TestBundleName       = `test-bundle`
	TestBundlePathPrefix = `/bundles/`
	BundleServiceKey     = "test-bundle-service"
)

// WithBundles is a test.Options that initialize OPA and start a bundle server in test with given bundle FS.
// All FSs are built into a single bundle and loaded to OPA engine.
// If no bundle FS provided, DefaultBundleFS is used.
func WithBundles(bundleFSs ...fs.FS) test.Options {
	return test.WithOptions(
		apptest.WithModules(opainit.Module),
		apptest.WithConfigFS(DefaultConfigFS),
		apptest.WithFxOptions(
			fx.Provide(BundleServerProvider(bundleFSs...)),
			fx.Invoke(opatestserver.InitializeBundleServer),
			fx.Invoke(WaitForOPAReady),
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

func WaitForOPAReady(lc fx.Lifecycle, ready opa.EmbeddedOPAReadyCH) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			select {
			case <-ready:
				return nil
			case <-ctx.Done():
				return fmt.Errorf("OPA Engine cannot be initialized before timeout")
			}
		},
	})
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
