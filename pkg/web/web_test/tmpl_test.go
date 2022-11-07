package web_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/assets"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/web_test/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"embed"
	"errors"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	gotemplate "html/template"
	"net/http"
	"regexp"
	"testing"
)

//go:embed testdata/tmpl/*
var TmplFS embed.FS

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

func SubTestStaticAssets(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		fsys := web.NewOSDirFS("testdata", web.DirFSAllowListDirectory)
		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			m := assets.New("/static", "testdata/").
				AddAlias("js", "static.js")
			e := reg.Register(fsys, m)
			g.Expect(e).To(Succeed(), "register assets should success")
		})
		testTextEndpoint(ctx, t, g, http.MethodGet, "/static/js", expectTextContentType("application/javascript"))
		testTextEndpoint(ctx, t, g, http.MethodGet, "/static/static.js", expectTextContentType("application/javascript"))
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
		mc.Set(web.ContextKeySession, "session-value")
		mc.Set(web.ContextKeySecurity, "security-value")
		mc.Set(web.ContextKeyCsrf, "csrf-value")
		testTextEndpoint(mc, t, g, http.MethodGet, "/index",
			expectTextContentType("text/html; charset=utf-8"),
			expectBodyModels(map[string]string{
				".Title":                  "TemplateMVCTest",
				".rc.ContextPath":         "/test",
				template.ModelKeySession:  "session-value",
				template.ModelKeySecurity: "security-value",
				template.ModelKeyCsrf:     "csrf-value",
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
		mc.Set(web.ContextKeySession, "session-value")
		mc.Set(web.ContextKeySecurity, "security-value")
		mc.Set(web.ContextKeyCsrf, "csrf-value")
		testTextEndpoint(mc, t, g, http.MethodGet, "/error",
			expectTextSC(http.StatusInternalServerError),
			expectTextContentType("text/html; charset=utf-8"),
			expectBodyModels(map[string]string{
				".rc.ContextPath":           "/test",
				template.ModelKeyError:      DefaultErrorMsg,
				template.ModelKeyMessage:    DefaultErrorMsg,
				template.ModelKeyStatusCode: "500",
				template.ModelKeyStatusText: http.StatusText(http.StatusInternalServerError),
				template.ModelKeySession:    "session-value",
				template.ModelKeySecurity:   "security-value",
				template.ModelKeyCsrf:       "csrf-value",
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
			pattern := fmt.Sprintf(testdata.ModelPrintTmpl, k, v)
			expect.body = append(expect.body, regexp.QuoteMeta(pattern))
		}
	}
}
