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

package actuator_tests

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/actuator"
	"github.com/cisco-open/go-lanai/pkg/actuator/actuator_tests/testdata"
	"github.com/cisco-open/go-lanai/pkg/actuator/alive"
	"github.com/cisco-open/go-lanai/pkg/actuator/info"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/actuatortest"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/sectest"
	. "github.com/cisco-open/go-lanai/test/utils/gomega"
	"github.com/cisco-open/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
	"testing"
)

/*************************
	Tests
 *************************/

// TestSimpleAdminEndpoints test simple endpoints like info, alive
func TestSimpleAdminEndpoints(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(webtest.AddDefaultRequestOptions(v3RequestOptions())),
		sectest.WithMockedMiddleware(),
		actuatortest.WithEndpoints(actuatortest.DisableAllEndpoints()),
		apptest.WithModules(info.Module, alive.Module),
		apptest.WithConfigFS(testdata.TestConfigFS),
		test.GomegaSubTest(SubTestInfoEndpointV3(), "TestInfoEndpointV3"),
		test.GomegaSubTest(SubTestInfoEndpointV2(), "TestInfoEndpointV2"),
		test.GomegaSubTest(SubTestAliveEndpoint(), "TestAliveEndpoint"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestInfoEndpointV3() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/info", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertInfoResponse(t, g, resp.Response)

		// By name, currently no supported
		//req = webtest.NewRequest(ctx, http.MethodGet, "/admin/info/app", nil)
		//resp = webtest.MustExec(ctx, req)
		//assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
	}
}

func SubTestInfoEndpointV2() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/info", nil, v2RequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV2)
		assertInfoResponse(t, g, resp.Response)

		// By name, currently no supported
		//req = webtest.NewRequest(ctx, http.MethodGet, "/admin/info/app", nil)
		//resp = webtest.MustExec(ctx, req)
		//assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
	}
}

func SubTestAliveEndpoint() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/alive", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
	}
}

/*************************
	Common Helpers
 *************************/

func assertInfoResponse(_ *testing.T, g *gomega.WithT, resp *http.Response) {
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `info response body should be readable`)
	g.Expect(body).To(HaveJsonPath("$.build-info.version"), "info response should build version")
}