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
	"github.com/cisco-open/go-lanai/pkg/actuator/env"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/actuatortest"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/sectest"
	"github.com/cisco-open/go-lanai/test/webtest"
	. "github.com/onsi/gomega"
	"net/http"
	"testing"
)

/*************************
	Tests
 *************************/

func TestEnvEndpoint(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(webtest.AddDefaultRequestOptions(v3RequestOptions())),
		sectest.WithMockedMiddleware(),
		actuatortest.WithEndpoints(actuatortest.DisableAllEndpoints()),
		apptest.WithModules(env.Module),
		apptest.WithConfigFS(testdata.TestConfigFS),
		test.GomegaSubTest(SubTestEnvWithAccess(mockedSecurityAdmin()), "TestEnvWithAccess"),
		test.GomegaSubTest(SubTestEnvWithoutAccess(mockedSecurityNonAdmin()), "TestEnvWithoutAccess"),
		test.GomegaSubTest(SubTestEnvWithoutAuth(), "TestEnvWithoutAuth"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestEnvWithAccess(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)
		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		actuatortest.AssertEnvResponse(t, resp.Response)
	}
}

func SubTestEnvWithoutAccess(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)

		// with non-admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusForbidden)
	}
}

func SubTestEnvWithoutAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusUnauthorized)
	}
}

/*************************
	Common Helpers
 *************************/


