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

package web_test

import (
    "context"
    "embed"
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/csrf"
    "github.com/cisco-open/go-lanai/pkg/security/session"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/assets"
    "github.com/cisco-open/go-lanai/pkg/web/template"
    "github.com/cisco-open/go-lanai/pkg/web/web_test/testdata"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/sectest"
    "github.com/cisco-open/go-lanai/test/webtest"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    gotemplate "html/template"
    "mime"
    "net/http"
    "regexp"
    "testing"
)

//go:embed testdata/tmpl/*
var TmplFS embed.FS

const (
	kModelTestRequest = "from-request"
	kModelTestStatic  = "static"
)

/*************************
	Tests
 *************************/

func TestTemplateRegistration(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithUtilities(),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(web.NewEngine),
		),
		test.Setup(SetupTestRegisterModelValuers()),
		test.SubTestSetup(ResetEngine(&di)),
		test.GomegaSubTest(SubTestStaticAssets(&di), "TestStaticAssets"),
		test.GomegaSubTest(SubTestCompressedAssets(&di), "TestCompressedAssets"),
		test.GomegaSubTest(SubTestTemplateMVC(&di), "TestTemplateMVC"),
		test.GomegaSubTest(SubTestTemplateError(&di), "TestTemplateError"),
		test.GomegaSubTest(SubTestTemplateRedirect(&di), "TestTemplateRedirect"),
	)
}

/*************************
	Sub Tests
 *************************/

func SetupTestRegisterModelValuers() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		template.RegisterGlobalModelValuer(kModelTestStatic, template.StaticModelValuer("static-value"))
		template.RegisterGlobalModelValuer(kModelTestRequest, template.RequestModelValuer(func(_ *http.Request) string {
			return "request-value"
		}))
		return ctx, nil
	}
}

func SubTestStaticAssets(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		fsys := web.NewOSDirFS("testdata", web.DirFSAllowListDirectory)
		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			m := assets.New("/static", "testdata/").
				AddAlias("js", "static.js")
			e := reg.Register(fsys, m)
			g.Expect(e).To(Succeed(), "register assets should success")
		})
		testTextEndpoint(ctx, t, g, http.MethodGet, "/static/js", expectTextContentType(mime.TypeByExtension(".js")))
		testTextEndpoint(ctx, t, g, http.MethodGet, "/static/static.js", expectTextContentType(mime.TypeByExtension(".js")))
	}
}

func SubTestCompressedAssets(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		fsys := web.NewOSDirFS("testdata", web.DirFSAllowListDirectory)
		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			m := assets.New("/static", "testdata/")
			e := reg.Register(fsys, m)
			g.Expect(e).To(Succeed(), "register assets should success")
		})
		resp := invokeEndpoint(ctx, t, g, http.MethodGet, "/static/compressed.js", webtest.Headers(
			"Accept-Encoding", "gzip",
		))
		expect := textExpectation{
			status: http.StatusOK,
			headers: map[string]string{
				"Content-Type":     "text/javascript",
				"Content-Encoding": "gzip",
			},
		}
		assertTextResponse(t, g, resp, expect)
	}
}

func SubTestTemplateMVC(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		fsys := web.NewOSDirFS("testdata", web.DirFSAllowListDirectory)
		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			m := template.Get("/index").HandlerFunc(testdata.IndexPage).Build()
			e := reg.Register(
				web.OrderedFS(TmplFS, 0),
				fsys, m)
			g.Expect(e).To(Succeed(), "register assets should success")

			e = reg.AddEngineOptions(func(eng *web.Engine) {
				eng.SetFuncMap(gotemplate.FuncMap{
					"printKV": testdata.PrintKV,
				})
			})
			g.Expect(e).To(Succeed(), "add template functions should success")
		})

		mc := utils.MakeMutableContext(ctx)
		security.MustSet(mc, security.EmptyAuthentication("security-value"))
		session.MustSet(mc, session.CreateSession(sectest.NewMockedSessionStore(), "session-value"))
		csrf.MustSet(mc, &csrf.Token{Value: "csrf-value"})
		testTextEndpoint(mc, t, g, http.MethodGet, "/index",
			expectTextContentType("text/html; charset=utf-8"),
			expectBodyModels(map[string]string{
				".Title":          "TemplateMVCTest",
				".rc.ContextPath": "/test",
				kModelTestStatic:  "static-value",
				kModelTestRequest: "request-value",
				"session":        `.+`,
				"security":       "security-value",
				"csrf.value":     "csrf-value",
			}),
		)
	}
}

func SubTestTemplateError(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		err := errors.New(DefaultErrorMsg)
		fsys := web.NewOSDirFS("testdata", web.DirFSAllowListDirectory)
		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			m := template.Get("/error").HandlerFunc(errorTmplEndpointFunc(err)).Build()
			e := reg.Register(web.OrderedFS(TmplFS, 0), fsys, m)
			g.Expect(e).To(Succeed(), "register assets should success")

			e = reg.AddEngineOptions(func(eng *web.Engine) {
				eng.SetFuncMap(gotemplate.FuncMap{
					"printKV": testdata.PrintKV,
				})
			})
			g.Expect(e).To(Succeed(), "add template functions should success")
		})

		mc := utils.MakeMutableContext(ctx)
		security.MustSet(mc, security.EmptyAuthentication("security-value"))
		session.MustSet(mc, session.CreateSession(sectest.NewMockedSessionStore(), "session-value"))
		csrf.MustSet(mc, &csrf.Token{Value: "csrf-value"})
		testTextEndpoint(mc, t, g, http.MethodGet, "/error",
			expectTextSC(http.StatusInternalServerError),
			expectTextContentType("text/html; charset=utf-8"),
			expectBodyModels(map[string]string{
				".rc.ContextPath":           "/test",
				template.ModelKeyError:      DefaultErrorMsg,
				template.ModelKeyMessage:    DefaultErrorMsg,
				template.ModelKeyStatusCode: "500",
				template.ModelKeyStatusText: http.StatusText(http.StatusInternalServerError),
				kModelTestStatic:            "static-value",
				kModelTestRequest:           "request-value",
				"session":                  `.+`,
				"security":                 "security-value",
				"csrf.value":               "csrf-value",
			}),
		)
	}
}

func SubTestTemplateRedirect(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		fsys := web.NewOSDirFS("testdata", web.DirFSAllowListDirectory)
		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			m := template.Get("/redirect").HandlerFunc(testdata.RedirectPage).Build()
			e := reg.Register(web.OrderedFS(TmplFS, 0), fsys, m)
			g.Expect(e).To(Succeed(), "register assets should success")
		})

		testTextEndpoint(ctx, t, g, http.MethodGet, "/redirect",
			expectTextSC(http.StatusFound),
			expectTextContentType("text/html; charset=utf-8"),
			expectTextHeader("Location", "/test/index"),
		)
	}
}

/*************************
	Helpers
 *************************/

func errorTmplEndpointFunc(err error) web.MvcHandlerFunc {
	return func(ctx context.Context, req *http.Request) (*template.ModelView, error) {
		return nil, err
	}
}

func expectTextSC(sc int) func(expect *textExpectation) {
	return func(expect *textExpectation) {
		expect.status = sc
	}
}

func expectTextContentType(contentType string) func(expect *textExpectation) {
	return func(expect *textExpectation) {
		expect.headers["Content-Type"] = contentType
	}
}

func expectTextHeader(k, v string) func(expect *textExpectation) {
	return func(expect *textExpectation) {
		expect.headers[k] = v
	}
}

func expectBodyModels(kvs map[string]string) func(expect *textExpectation) {
	return func(expect *textExpectation) {
		expect.body = make([]string, 0, len(kvs))
		for k, v := range kvs {
			pattern := fmt.Sprintf(testdata.ModelPrintTmpl, regexp.QuoteMeta(k), v)
			expect.body = append(expect.body, pattern)
		}
	}
}
