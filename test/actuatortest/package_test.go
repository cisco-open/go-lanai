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

package actuatortest

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/actuator"
	"github.com/cisco-open/go-lanai/pkg/actuator/apilist"
	"github.com/cisco-open/go-lanai/pkg/actuator/env"
	"github.com/cisco-open/go-lanai/pkg/actuator/health"
	healthep "github.com/cisco-open/go-lanai/pkg/actuator/health/endpoint"
	"github.com/cisco-open/go-lanai/pkg/actuator/loggers"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/actuatortest/testdata"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/sectest"
	"github.com/cisco-open/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

/*************************
	Test Setup
 *************************/

func ConfigureHealth(healthReg health.Registrar, mock *testdata.MockedHealthIndicator) {
	healthReg.MustRegister(mock)
}

/*************************
	Tests
 *************************/

func TestActuatorEndpoints(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(
			webtest.AddDefaultRequestOptions(webtest.Headers("Accept", "application/json")),
		),
		sectest.WithMockedMiddleware(),
		WithEndpoints(DisableAllEndpoints()),
		apptest.WithModules(
			health.Module, healthep.Module, loggers.Module, env.Module, apilist.Module,
		),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithFxOptions(
			fx.Provide(testdata.NewMockedHealthIndicator),
			fx.Invoke(ConfigureHealth),
		),
		apptest.WithDI(),
		test.GomegaSubTest(SubTestHealthWithDetails(), "TestHealthWithDetails"),
		test.GomegaSubTest(SubTestLoggersEndpoint(), "TestLoggersEndpoint"),
		test.GomegaSubTest(SubTestEnvEndpoint(), "TestEnvEndpoint"),
		test.GomegaSubTest(SubTestAPIListEndpoint(), "TestAPIListEndpoint"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestHealthWithDetails() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertHealthResponse(t, resp.Response, ExpectHealthDetails(), ExpectHealthComponents("test"))

		// with admin security GET V2
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, v2RequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV2)
		AssertHealthResponse(t, resp.Response, ExpectHealthDetails(), ExpectHealthComponents("test"))
	}
}

func SubTestLoggersEndpoint() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertLoggersResponse(t, resp.Response)

		// with admin security GET with name
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers/bootstrap", nil)
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertLoggersResponse(t, resp.Response, ExpectLoggersSingleEntry())
	}
}

func SubTestAPIListEndpoint() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/apilist", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertAPIListResponse(t, resp.Response)
	}
}

func SubTestEnvEndpoint() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertEnvResponse(t, resp.Response)
	}
}

/*************************
	Helpers
 *************************/

func v2RequestOptions() webtest.RequestOptions {
	return func(req *http.Request) {
		req.Header.Set("Accept", actuator.ContentTypeSpringBootV2)
	}
}

func assertResponse(_ *testing.T, g *gomega.WithT, resp *http.Response, expectedStatus int, expectedHeaders ...string) {
	g.Expect(resp).ToNot(BeNil(), "endpoint should have response")
	g.Expect(resp.StatusCode).To(BeEquivalentTo(expectedStatus))
	for i := range expectedHeaders {
		if i%2 == 1 || i+1 >= len(expectedHeaders) {
			continue
		}
		k := expectedHeaders[i]
		v := expectedHeaders[i+1]
		g.Expect(resp.Header.Get(k)).To(BeEquivalentTo(v), "response header should contains [%s]='%s'", k, v)
	}
}