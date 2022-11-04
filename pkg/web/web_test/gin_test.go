package web_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/web_test/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

/*************************
	Tests
 *************************/

func TestGinMiddlewares(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithUtilities(),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(web.NewEngine),
		),
		test.SubTestSetup(ResetEngine(&di)),
		test.GomegaSubTest(SubTestGinContextAvailability(&di), "TestGinContextAvailability"),
		test.GomegaSubTest(SubTestContextKV(&di), "TestContextKV"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestGinContextAvailability(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		assertion := func(ctx context.Context, req *http.Request) {
			g.Expect(web.GinContext(ctx)).To(Not(BeNil()), "gin.Context from ctx should not be nil")
			g.Expect(web.GinContext(req.Context())).To(Not(BeNil()), "gin.Context from ctx should not be nil")
			g.Expect(web.HttpRequest(ctx)).To(Equal(req), "web.HttpRequest should return same request")

			var e error
			var ret interface{}
			e = withRecover(func() error { ret = web.MustGinContext(ctx); return nil })
			g.Expect(e).To(Succeed(), "MustGinContext shouldn't panic")
			g.Expect(ret).To(Not(BeNil()), "gin.Context from ctx should not be nil")

			e = withRecover(func() error { ret = web.MustHttpRequest(ctx); return nil })
			g.Expect(e).To(Succeed(), "MustHttpRequest shouldn't panic")
			g.Expect(ret).To(Equal(req), "web.MustHttpRequest should return same request")
		}

		// execute test
		WebInit(ctx, t, g, di,
			registerAssertingEndpoint(http.MethodPost, "/mw/:var", assertion),
			registerAssertingMW(http.MethodPost, "/mw/**", assertion),
		)
		testEndpoint(ctx, t, g, http.MethodPost, "/mw/var-value")
	}
}

type kv struct {
	k string
	v string
}

func SubTestContextKV(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		kvs := map[kv]kvSetter{
			kv{k: "gin-ctx", v: "gin-ctx-value"}:         ginCtxKVSetter(),
			kv{k: "req-gin-ctx", v: "req-gin-ctx-value"}: reqGinCtxKVSetter(),
			kv{k: "web-ctx", v: "web-ctx-value"}:         webCtxKVSetter(),
			kv{k: "web-req", v: "web-req-value"}:         webReqKVSetter(),
			//kv{k: "mutable-ctx", v: "mutable-ctx"}:       mutableCtxKVSetter(),
		}
		mwAssertion := func(ctx context.Context, req *http.Request) {
			for kv, setter := range kvs {
				setter(ctx, req, kv.k, kv.v)
			}
		}

		epAssertion := func(ctx context.Context, req *http.Request) {
			for kv := range kvs {
				g.Expect(ctx.Value(kv.k)).To(Equal(kv.v), "context should contains correct KV: %s=%s", kv.k, kv.v)
				g.Expect(req.Context().Value(kv.k)).To(Equal(kv.v), "Request context should contains correct KV: %s=%s", kv.k, kv.v)
			}
		}

		// execute test
		WebInit(ctx, t, g, di,
			registerAssertingEndpoint(http.MethodPost, "/mw/:var", epAssertion),
			registerAssertingMW(http.MethodPost, "/mw/**", mwAssertion),
		)
		testEndpoint(ctx, t, g, http.MethodPost, "/mw/var-value")
	}
}

/*************************
	Helper
 *************************/

func registerAssertingEndpoint(method, path string, fn assertionFunc) WebInitFunc {
	return func(reg *web.Registrar) {
		reg.MustRegister(rest.New(path).
			Method(method).
			Path(path).
			EndpointFunc(assertingEndpointFunc(fn)).
			Build())
	}
}

func registerAssertingMW(method, pattern string, fn assertionFunc) WebInitFunc {
	return func(reg *web.Registrar) {
		reg.MustRegister(middleware.NewBuilder("mw").
			ApplyTo(matcher.RouteWithPattern(pattern, method)).
			Use(assertingMWFunc(fn)).
			Build())
	}
}

type assertionFunc func(ctx context.Context, req *http.Request)

func assertingEndpointFunc(assertFn assertionFunc) web.MvcHandlerFunc {
	return func(ctx context.Context, req *http.Request) (interface{}, error) {
		assertFn(ctx, req)
		return testdata.Raw(ctx, req)
	}
}

func assertingMWFunc(assertFn assertionFunc) web.HandlerFunc {
	return func(_ http.ResponseWriter, req *http.Request) {
		assertFn(req.Context(), req)
	}
}

type kvSetter func(ctx context.Context, req *http.Request, k, v string)

func ginCtxKVSetter() kvSetter {
	return func(ctx context.Context, req *http.Request, k, v string) {
		gc := web.GinContext(ctx)
		gc.Set(k, v)
	}
}

func reqGinCtxKVSetter() kvSetter {
	return func(ctx context.Context, req *http.Request, k, v string) {
		gc := web.GinContext(req.Context())
		gc.Set(k, v)
	}
}

func webCtxKVSetter() kvSetter {
	return func(ctx context.Context, req *http.Request, k, v string) {
		web.SetKV(ctx, k, v)
	}
}

func webReqKVSetter() kvSetter {
	return func(ctx context.Context, req *http.Request, k, v string) {
		web.SetKV(req.Context(), k, v)
	}
}

func mutableCtxKVSetter() kvSetter {
	return func(ctx context.Context, req *http.Request, k, v string) {
		// TODO
	}
}
