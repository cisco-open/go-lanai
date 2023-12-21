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

package swagger

import (
	"context"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"embed"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
	"testing"
)

/*************************
	Setup
 *************************/

//go:embed testdata/*
var testFS embed.FS

func initializeTest(di initDI) {
	di.Registrar.MustRegister(Content)
	di.Registrar.MustRegister(newSwaggerController(di.Properties, di.Resolver, testFS))
}

/*************************
	Tests
 *************************/

type swaggerDI struct {
	fx.In
	Registrar *web.Registrar
}

func TestOAS3Endpoints(t *testing.T) {
	di := &swaggerDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithDI(di),
		apptest.WithProperties(
			"swagger.spec: testdata/api-docs-v3.yml", "swagger.base-path: /myapp",
		),
		apptest.WithFxOptions(
			appconfig.FxEmbeddedDefaults(defaultConfigFS),
			fx.Provide(bindSwaggerProperties),
			fx.Invoke(initializeTest),
		),
		test.GomegaSubTest(SubTestSwaggerUIPage(), "TestSwaggerUIPage"),
		test.GomegaSubTest(SubTestSwaggerRedirectPage(), "TestSwaggerRedirectPage"),
		test.GomegaSubTest(SubTestUIConfiguration(), "TestUIConfiguration"),
		test.GomegaSubTest(SubTestSecurityConfiguration(), "TestSecurityConfiguration"),
		test.GomegaSubTest(SubTestSsoConfiguration(), "TestSsoConfiguration"),
		test.GomegaSubTest(SubTestResourcesConfiguration(), "TestResourcesConfiguration"),
		test.GomegaSubTest(SubTestOAS3DocsNoServer(), "TestOAS3DocsNoServer"),
		test.GomegaSubTest(SubTestOAS3DocsForwardHostOnly(), "TestOAS3DocsForwardHostOnly"),
		test.GomegaSubTest(SubTestOAS3DocsForwardHostAndProto(), "TestOAS3DocsForwardHostAndProto"),
	)
}

func TestOAS3EndpointsWithoutBasePath(t *testing.T) {
	di := &swaggerDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithDI(di),
		apptest.WithProperties(
			"swagger.spec: testdata/api-docs-v3.yml",
		),
		apptest.WithFxOptions(
			appconfig.FxEmbeddedDefaults(defaultConfigFS),
			fx.Provide(bindSwaggerProperties),
			fx.Invoke(initializeTest),
		),
		test.GomegaSubTest(SubTestOAS3DocsForwardHostAndProtoWithoutBasePath(), "TestOAS3DocsForwardHostAndProtoWithoutBasePath"),
	)
}

func TestOAS3EndpointsBasePathWithoutSlash(t *testing.T) {
	di := &swaggerDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithDI(di),
		apptest.WithProperties(
			"swagger.spec: testdata/api-docs-v3.yml", "swagger.base-path: myapp",
		),
		apptest.WithFxOptions(
			appconfig.FxEmbeddedDefaults(defaultConfigFS),
			fx.Provide(bindSwaggerProperties),
			fx.Invoke(initializeTest),
		),
		test.GomegaSubTest(SubTestOAS3DocsForwardHostAndProto(), "TestOAS3DocsForwardHostAndProtoAndBasePathWithoutSlash"),
	)
}

func TestOAS2Endpoints(t *testing.T) {
	di := &swaggerDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithDI(di),
		apptest.WithProperties(
			"swagger.spec: testdata/api-docs-v2.json",
		),
		apptest.WithFxOptions(
			appconfig.FxEmbeddedDefaults(defaultConfigFS),
			fx.Provide(bindSwaggerProperties),
			fx.Invoke(initializeTest),
		),
		test.GomegaSubTest(SubTestOAS2DocsNoServer("v2"), "TestOAS2DocsV2ApiNoServer"),
		test.GomegaSubTest(SubTestOAS2DocsNoServer("v3"), "TestOAS2DocsV3ApiNoServer"),
		test.GomegaSubTest(SubTestOAS2DocsForwardHost("v2"), "TestOAS2DocsV2ApiForwardHost"),
		test.GomegaSubTest(SubTestOAS2DocsForwardHost("v3"), "TestOAS2DocsV3ApiForwardHost"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestSwaggerUIPage() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var r *http.Request
		var resp *http.Response
		r = webtest.NewRequest(ctx, http.MethodGet, "/swagger", nil)
		resp = exec(ctx, r)
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		contentType := resp.Header.Get("Content-Type")
		g.Expect(contentType).To(HavePrefix("text/html"), "UI page should have text/html content type")

		// with slash
		r = webtest.NewRequest(ctx, http.MethodGet, "/swagger/", nil)
		resp = exec(ctx, r)
		g.Expect(resp.StatusCode).To(Equal(http.StatusMovedPermanently), "StatusCode should be correct")
		contentType = resp.Header.Get("Content-Type")
		g.Expect(contentType).To(HavePrefix("text/html"), "UI page should have text/html content type")
		g.Expect(resp.Header.Get("Location")).To(Equal("http:///test/swagger"))
	}
}

func SubTestSwaggerRedirectPage() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var r *http.Request
		var resp *http.Response
		r = webtest.NewRequest(ctx, http.MethodGet, "swagger-sso-redirect.html", nil)
		resp = exec(ctx, r)
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		contentType := resp.Header.Get("Content-Type")
		g.Expect(contentType).To(HavePrefix("text/html"), "Redirect page should have text/html content type")
	}
}

func SubTestUIConfiguration() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var r *http.Request
		var resp *http.Response
		r = webtest.NewRequest(ctx, http.MethodGet, "/swagger-resources/configuration/ui", nil)
		resp = exec(ctx, r)
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_ui_config.json")
	}
}

func SubTestSecurityConfiguration() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var r *http.Request
		var resp *http.Response
		r = webtest.NewRequest(ctx, http.MethodGet, "/swagger-resources/configuration/security", nil)
		resp = exec(ctx, r)
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_security_config.json")
	}
}

func SubTestSsoConfiguration() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var r *http.Request
		var resp *http.Response
		r = webtest.NewRequest(ctx, http.MethodGet, "/swagger-resources/configuration/security/sso", nil)
		resp = exec(ctx, r)
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_sso_config.json")
	}
}

func SubTestResourcesConfiguration() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var r *http.Request
		var resp *http.Response
		r = webtest.NewRequest(ctx, http.MethodGet, "/swagger-resources", nil)
		resp = exec(ctx, r)
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_resources_config.json")
	}
}

func SubTestOAS3DocsNoServer() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var r *http.Request
		var resp *http.Response
		r = webtest.NewRequest(ctx, http.MethodGet, "/v3/api-docs", nil)
		resp = exec(ctx, r)
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_oas3_no_server.json")
	}
}

func SubTestOAS3DocsForwardHostOnly() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var r *http.Request
		var resp *http.Response

		// single host
		r = webtest.NewRequest(ctx, http.MethodGet, "/v3/api-docs", nil)
		resp = exec(ctx, r, withHeader("X-Forwarded-Host", "my.test.server:9876"))
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_oas3_http_server.json")

		// multiple hosts
		r = webtest.NewRequest(ctx, http.MethodGet, "/v3/api-docs", nil)
		resp = exec(ctx, r, withHeader("X-Forwarded-Host", "my.test.server:9876,my.alt.test.server:8765"))
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_oas3_http_server.json")
	}
}

func SubTestOAS3DocsForwardHostAndProto() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var r *http.Request
		var resp *http.Response

		// single host
		r = webtest.NewRequest(ctx, http.MethodGet, "/v3/api-docs", nil)
		resp = exec(ctx, r,
			withHeader("X-Forwarded-Host", "my.test.server:9876"),
			withHeader("X-Forwarded-Proto", "https"),
		)
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_oas3_https_server.json")

		// multiple hosts
		r = webtest.NewRequest(ctx, http.MethodGet, "/v3/api-docs", nil)
		resp = exec(ctx, r,
			withHeader("X-Forwarded-Host", "my.test.server:9876,my.alt.test.server:8765"),
			withHeader("X-Forwarded-Proto", "https,https"),
		)
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_oas3_https_server.json")
	}
}

func SubTestOAS3DocsForwardHostAndProtoWithoutBasePath() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var r *http.Request
		var resp *http.Response

		// single host
		r = webtest.NewRequest(ctx, http.MethodGet, "/v3/api-docs", nil)
		resp = exec(ctx, r,
			withHeader("X-Forwarded-Host", "my.test.server:9876"),
			withHeader("X-Forwarded-Proto", "https"),
		)
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_oas3_https_server_no_base_path.json")

		// multiple hosts
		r = webtest.NewRequest(ctx, http.MethodGet, "/v3/api-docs", nil)
		resp = exec(ctx, r,
			withHeader("X-Forwarded-Host", "my.test.server:9876,my.alt.test.server:8765"),
			withHeader("X-Forwarded-Proto", "https,https"),
		)
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_oas3_https_server_no_base_path.json")
	}
}

func SubTestOAS2DocsNoServer(apiVer string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var r *http.Request
		var resp *http.Response
		r = webtest.NewRequest(ctx, http.MethodGet, fmt.Sprintf("/%s/api-docs", apiVer), nil)
		resp = exec(ctx, r)
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_oas2_no_server.json")
	}
}

func SubTestOAS2DocsForwardHost(apiVer string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var r *http.Request
		var resp *http.Response

		// single host
		r = webtest.NewRequest(ctx, http.MethodGet, fmt.Sprintf("/%s/api-docs", apiVer), nil)
		resp = exec(ctx, r, withHeader("X-Forwarded-Host", "my.test.server:9876"))
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_oas2_with_forward.json")

		// multiple hosts
		r = webtest.NewRequest(ctx, http.MethodGet, fmt.Sprintf("/%s/api-docs", apiVer), nil)
		resp = exec(ctx, r, withHeader("X-Forwarded-Host", "my.test.server:9876,my.alt.test.server:8765"))
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "StatusCode should be correct")
		assertJsonBody(t, g, resp, "testdata/expected_oas2_with_forward.json")
	}
}

/*************************
	Helpers
 *************************/

type mockedReqOpts func(*http.Request)

func withHeader(k, v string) mockedReqOpts {
	return func(r *http.Request) {
		r.Header.Set(k, v)
	}
}

func exec(ctx context.Context, r *http.Request, opts ...mockedReqOpts) *http.Response {
	r.Header.Set("Content-Type", "application/json")
	for _, fn := range opts {
		fn(r)
	}
	return webtest.MustExec(ctx, r).Response
}

func assertJsonBody(t *testing.T, g *gomega.WithT, resp *http.Response, expectedJsonFile string) {
	// read data
	data, e := testFS.ReadFile(expectedJsonFile)
	g.Expect(e).To(Succeed(), "should able to read data JSON file '%s'", expectedJsonFile)
	expected := string(data)
	// check body
	body := readBodyAsString(t, g, resp)
	g.Expect(body).To(MatchJSON(expected), "response body should be a JSON matching data values")
}

func readBodyAsString(_ *testing.T, g *gomega.WithT, resp *http.Response) string {
	defer func() { _ = resp.Body.Close() }()
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), "parsing body should not returns error")
	return string(body)
}
