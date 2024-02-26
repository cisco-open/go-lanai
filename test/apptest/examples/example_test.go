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

package examples

import (
    "context"
    "embed"
    "github.com/cisco-open/go-lanai/pkg/data/tx"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/rest"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/dbtest"
    "github.com/cisco-open/go-lanai/test/webtest"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "net/http"
    "testing"
    "time"
)

/*************************
	Setup
 *************************/

type DummyService interface {
	DummyMethod(_ context.Context) error
}

func NewRealService() DummyService {
	return &realService{}
}

type realService struct {}

func (t *realService) DummyMethod(_ context.Context) error {
	return nil
}

type DummyController struct{
	svc DummyService
}

func NewDummyController(svc DummyService) web.Controller {
	return &DummyController{
		svc: svc,
	}
}

func (c *DummyController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.New("test").Get("/api").EndpointFunc(c.Test).Build(),
	}
}

func (c *DummyController) Test(_ context.Context, _ *http.Request) (response interface{}, err error) {
	return map[string]string{
		"message": "ok",
	}, nil
}

type serviceDI struct {
	fx.In
	Service DummyService
}

/*************************
	Examples
 *************************/

// TestBootstrapWithDefaults
// Simple configuration to setup a set of sub tests with FX and bootstrapping
// Any number of DI struct pointers (with fx.In) can be specified.
// All specified DI struct via apptest.WithDI will be populated when sub tests is run
// Only bootstrap and appconfig package is initialized by default
func TestBootstrapWithDefaults(t *testing.T) {
	di := serviceDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(&di), 	// tell test framework to do dependencies injection
		apptest.WithFxOptions(
			fx.Provide(NewRealService), // provide real service
		),
		test.GomegaSubTest(SubTestExampleWithRealService(&di), "SubTestWithRealService"),
	)
}

//go:embed example-test-config.yml
var customConfigFS embed.FS

// TestBootstrapWithCustomConfigAndMocks
// Mocked services can be provided via apptest.WithFxOptions.
// But due to limitation of uber.fx, no provider overwriting is supported
// Config FS is loaded as application ad-hoc configs, which has higher priority than all but CLI/consul/vault config
func TestBootstrapWithCustomConfigAndMocks(t *testing.T) {
	di := serviceDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(&di), 	// tell test framework to do dependencies injection
		apptest.WithTimeout(30*time.Second),
		apptest.WithConfigFS(customConfigFS),
		apptest.WithProperties("info.inline: value", "info.inline-alt=value"),
		apptest.WithFxOptions(
			fx.Provide(NewMockedService), // provide real service
		),
		test.GomegaSubTest(SubTestExampleWithMockedService(&di), "SubTestWithMockedService"),
	)
}

// TestBootstrapWithRealWebServer
// when using together with webtest.WithRealServer(), the web package would create a real server with random port.
// The port can be retrieved via *web.Registrar, or webtest.CurrentPort()
// webtest.NewRequest() is also available to automatically fill in host, port and context-path
func TestBootstrapWithRealWebServer(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		apptest.WithFxOptions(
			// provide DummyController and mocked service
			fx.Provide(NewMockedService),
			web.FxControllerProviders(NewDummyController),
		),
		test.GomegaSubTest(SubTestExampleWithRealWebController(), "SubTestWithRealWebController"),
	)
}

// TestBootstrapWithMockedWebServer
// when using together with webtest.WithMockedServer(), the web package would initialize all components as usual
// without creating a real web server, which would save TCP resources and be faster when execute.
// In this mode, webtest.NewRequest() and webtest.Exec() (or webtest.MustExec()) should be used to create and execute request
func TestBootstrapWithMockedWebServer(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithFxOptions(
			// provide DummyController and mocked service
			fx.Provide(NewMockedService),
			web.FxControllerProviders(NewDummyController),
		),
		test.GomegaSubTest(SubTestExampleWithRealWebController(), "SubTestWithRealWebController"),
	)
}

// TestBootstrapWithDataMocks
// Without specifying server port when webinit.Module is enabled, the web package would create a real
// server with random port. The port can be retrieved via *web.Registrar
// Note: due to current bootstrapping limitation, all modules added via apptest.WithModules would affect other tests,
//		 so no conflicting modules is allowed between all tests of same package.
func TestBootstrapWithDataMocks(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithNoopMocks(),
		test.GomegaSubTest(SubTestExampleWithOverriddenTxManager(), "SubTestWithOverriddenTxManager"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestExampleWithRealService(di *serviceDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Service).To(Not(BeNil()), "Service should be injected")
		g.Expect(di.Service).To(BeAssignableToTypeOf(&realService{}), "Injected service should be the real service")
	}
}

func SubTestExampleWithMockedService(di *serviceDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Service).To(Not(BeNil()), "Service should be injected")
		g.Expect(di.Service).To(BeAssignableToTypeOf(&mockedService{}), "Injected service should be the real service")
	}
}

func SubTestExampleWithRealWebController() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		req := webtest.NewRequest(ctx, http.MethodGet,"/api", nil)
		// Alternatively: resp := webtest.MustExec(ctx, req)
		ret, e := webtest.Exec(ctx, req)
		g.Expect(e).To(Succeed(), "http client should be succeeded")
		resp := ret.Response
		g.Expect(resp).To(Not(BeNil()), "http response should not be nil ")
		g.Expect(resp.StatusCode).To(Equal(200), "http response status should be 200")
	}
}

func SubTestExampleWithOverriddenTxManager() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// Regular usage
		e := tx.Transaction(ctx, func(txCtx context.Context) error {
			_, ok := txCtx.(tx.TxContext)
			g.Expect(ok).To(BeTrue(), "Overridden TxManager should create tx.TxContext")
			_, ok = txCtx.(tx.GormContext)
			g.Expect(ok).To(BeTrue(), "Overridden TxManager should create tx.GormContext")
			return nil
		})
		g.Expect(e).To(Succeed(), "Overridden TxManager shouldn't return error")

		// Manual usage
		txCtx, e := tx.Begin(ctx)
		g.Expect(e).To(Succeed(), "Overridden ManualTxManager shouldn't return error")
		_, ok := txCtx.(tx.TxContext)
		g.Expect(ok).To(BeTrue(), "Overridden TxManager should create tx.TxContext")
		_, ok = txCtx.(tx.GormContext)
		g.Expect(ok).To(BeTrue(), "Overridden TxManager should create tx.GormContext")

		txCtx, e = tx.Commit(txCtx)
		g.Expect(e).To(Succeed(), "Overridden ManualTxManager shouldn't return error")
		g.Expect(txCtx).To(BeIdenticalTo(ctx), "Overridden ManualTxManager shouldn't do anything")
	}
}