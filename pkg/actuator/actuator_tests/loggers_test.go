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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/actuator_tests/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/loggers"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/actuatortest"
	. "cto-github.cisco.com/NFV-BU/go-lanai/test/actuatortest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"fmt"
	. "github.com/onsi/gomega"
	"net/http"
	"strings"
	"testing"
)

var _ = log.New("Test")

/*************************
	Tests
 *************************/

func TestLoggersEndpoint(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(webtest.AddDefaultRequestOptions(v3RequestOptions())),
		sectest.WithMockedMiddleware(),
		actuatortest.WithEndpoints(actuatortest.DisableAllEndpoints()),
		apptest.WithModules(loggers.Module),
		apptest.WithConfigFS(testdata.TestConfigFS),
		test.GomegaSubTest(SubTestLoggersWithAccess(mockedSecurityAdmin()), "TestLoggersWithAccess"),
		test.GomegaSubTest(SubTestLoggersWithoutAccess(mockedSecurityNonAdmin()), "TestLoggersWithoutAccess"),
		test.GomegaSubTest(SubTestLoggersWithoutAuth(), "TestLoggersWithoutAuth"),
		test.GomegaSubTest(SubTestChangeLoggers(mockedSecurityAdmin()), "TestChangeLoggers"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestLoggersWithAccess(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)
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

func SubTestLoggersWithoutAccess(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)

		// with non-admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusForbidden)

		// with non-admin security GET with name
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers/test", nil)
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusForbidden)
	}
}

func SubTestLoggersWithoutAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusUnauthorized)

		// regular GET with name
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers/test", nil)
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusUnauthorized)
	}
}

func SubTestChangeLoggers(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)
		// send POST
		const newLevel = "ERROR"
		body := fmt.Sprintf(`{"configuredLevel":"%s"}`, newLevel)
		req := webtest.NewRequest(ctx, http.MethodPost, "/admin/loggers/test", strings.NewReader(body),
			webtest.ContentType("application/json"))
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusNoContent)

		// check Result
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers/test", nil)
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertLoggersResponse(t, resp.Response, ExpectLoggersSingleEntry(newLevel, newLevel))
	}
}

/*************************
	Common Helpers
 *************************/



