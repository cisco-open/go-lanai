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

package opaaccess_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	opaaccess "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/access"
	opatest "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
	"time"
)

/************************
	Test Setup
 ************************/

func RegisterSecurityConfigurer(reg security.Registrar) {
	configurer := security.ConfigurerFunc(func(ws security.WebSecurity) {
		ws.Route(matcher.RouteWithPattern("/api/**")).
			With(access.New().
				Request(matcher.AnyRequest()).
				WithOrder(order.Highest).
				CustomDecisionMaker(opaaccess.DecisionMakerWithOPA(opa.RequestQueryWithPolicy("testservice/allow_api"))),
			).
			With(errorhandling.New())
	})
	reg.Register(configurer)
}

func ProvideTestController() web.Controller {
	return TestController{}
}

type TestController struct {}

func (c TestController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.Get("/api/get").EndpointFunc(c.Get).Build(),
		rest.Post("/api/post").EndpointFunc(c.Post).Build(),
		rest.Delete("/api/delete").EndpointFunc(c.Delete).Build(),
	}
}
func (c TestController) Get(_ context.Context, _ *http.Request) (interface{}, error) {
	return "ok", nil
}

func (c TestController) Post(_ context.Context, _ *http.Request) (int, interface{}, error) {
	return http.StatusCreated, nil, nil
}

func (c TestController) Delete(_ context.Context, _ *http.Request) (int, interface{}, error) {
	return http.StatusNoContent, nil, nil
}

/************************
	Tests
 ************************/

type TestDI struct {
	fx.In
}

func TestWithMockedServer(t *testing.T) {
	di := &TestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithTimeout(10 * time.Minute),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(),
		opatest.WithBundles(),
		apptest.WithModules(access.Module, errorhandling.Module),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			web.FxControllerProviders(ProvideTestController),
			fx.Invoke(RegisterSecurityConfigurer),
		),
		test.GomegaSubTest(SubTestGet(di), "TestGet"),
		test.GomegaSubTest(SubTestPost(di), "TestPost"),
		test.GomegaSubTest(SubTestDelete(di), "TestDelete"),
	)
}

/************************
	SubTests
 ************************/

func SubTestGet(_ *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		// admin
		ctx = sectest.ContextWithSecurity(ctx, AdminOptions())
		req = webtest.NewRequest(ctx, http.MethodGet, "/api/get", nil, webtest.ContentType("application/json"))
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "admin should be able to GET")

		// viewer
		ctx = sectest.ContextWithSecurity(ctx, ViewerOptions())
		req = webtest.NewRequest(ctx, http.MethodGet, "/api/get", nil, webtest.ContentType("application/json"))
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "viewer should be able to GET")

		// manager
		ctx = sectest.ContextWithSecurity(ctx, ManagerOptions())
		req = webtest.NewRequest(ctx, http.MethodGet, "/api/get", nil, webtest.ContentType("application/json"))
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "manager should be able to GET")
	}
}

func SubTestPost(_ *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		// admin
		ctx = sectest.ContextWithSecurity(ctx, AdminOptions())
		req = webtest.NewRequest(ctx, http.MethodPost, "/api/post", nil, webtest.ContentType("application/json"))
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp.StatusCode).To(Equal(http.StatusCreated), "admin should be able to POST")

		// viewer
		ctx = sectest.ContextWithSecurity(ctx, ViewerOptions())
		req = webtest.NewRequest(ctx, http.MethodPost, "/api/post", nil, webtest.ContentType("application/json"))
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp.StatusCode).To(Equal(http.StatusForbidden), "viewer should not be able to POST")

		// manager
		ctx = sectest.ContextWithSecurity(ctx, ManagerOptions())
		req = webtest.NewRequest(ctx, http.MethodPost, "/api/post", nil, webtest.ContentType("application/json"))
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp.StatusCode).To(Equal(http.StatusCreated), "manager should be able to POST")
	}
}

func SubTestDelete(_ *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// NOTE: DELETE permissions is not defined the default policies, no one should access it
		var req *http.Request
		var resp *http.Response
		// admin
		ctx = sectest.ContextWithSecurity(ctx, AdminOptions())
		req = webtest.NewRequest(ctx, http.MethodDelete, "/api/delete", nil, webtest.ContentType("application/json"))
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp.StatusCode).To(Equal(http.StatusForbidden), "admin should not be able to DELETE")

		// viewer
		ctx = sectest.ContextWithSecurity(ctx, ViewerOptions())
		req = webtest.NewRequest(ctx, http.MethodDelete, "/api/delete", nil, webtest.ContentType("application/json"))
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp.StatusCode).To(Equal(http.StatusForbidden), "viewer should not be able to DELETE")

		// manager
		ctx = sectest.ContextWithSecurity(ctx, ManagerOptions())
		req = webtest.NewRequest(ctx, http.MethodDelete, "/api/delete", nil, webtest.ContentType("application/json"))
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp.StatusCode).To(Equal(http.StatusForbidden), "manager should not be able to DELETE")
	}
}

/************************
	Helpers
 ************************/

func CommonMockOptions(d *sectest.SecurityDetailsMock) {
	d.UserId = uuid.New().String()
	d.TenantId = uuid.New().String()
	d.ProviderId = uuid.New().String()
	d.Tenants = utils.NewStringSet(d.TenantId)
}

func AdminOptions() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		CommonMockOptions(d)
		d.Username = "admin"
		d.Permissions = utils.NewStringSet("ADMIN")
		d.Roles = utils.NewStringSet("ADMIN")
	})
}

func ViewerOptions() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		CommonMockOptions(d)
		d.Username = "viewer"
		d.Permissions = utils.NewStringSet("VIEW", "NO_MANAGE")
		d.Roles = utils.NewStringSet("VIEWER")
	})
}

func ManagerOptions() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		CommonMockOptions(d)
		d.Username = "manager"
		d.Permissions = utils.NewStringSet("MANAGE", "VIEW")
		d.Roles = utils.NewStringSet("MANAGER")
	})
}

