package web_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/web_test/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"regexp"
	"strconv"
	"testing"
)

/*************************
	Tests
 *************************/

var logger = log.New("Web.Test")

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
		test.GomegaSubTest(SubTestContextDefaultKV(&di), "TestContextDefaultKV"),
		test.GomegaSubTest(SubTestContextSetKV(&di), "TestContextSetKV"),
		test.GomegaSubTest(SubTestGinHandlerMapping(&di), "TestGinHandlerMapping"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestGinContextAvailability(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		assertion := func(ctx context.Context, req *http.Request) {
			assertContext(ctx, t, g)
			g.Expect(web.GinContext(req.Context())).To(Not(BeNil()), "gin.Context from ctx should not be nil")
			g.Expect(web.HttpRequest(ctx)).To(Equal(req), "web.HttpRequest should return same request")

			var e error
			var ret interface{}
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

func SubTestContextDefaultKV(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		kvs := []kv{
			makeKV(web.ContextKeyContextPath, webtest.DefaultContextPath, kvSrcAll),
			makeKV(web.ContextKeyContextPath, webtest.CurrentContextPath(ctx), kvSrcAll),
		}

		assertion := func(ctx context.Context, req *http.Request) {
			for _, kv := range kvs {
				g.Expect(ctx.Value(kv.k)).To(Equal(kv.v), "context should contain correct K%s=%s", kv.k, kv.v)
				g.Expect(req.Context().Value(kv.k)).To(Equal(kv.v), "Request context should contain correct K%s=%s", kv.k, kv.v)
			}
		}

		// execute test
		WebInit(ctx, t, g, di,
			registerAssertingEndpoint(http.MethodPost, "/mw/:var", assertion),
			registerAssertingMW(http.MethodPost, "/mw/**", assertion),
		)
		testEndpoint(ctx, t, g, http.MethodPost, "/mw/var-value")
	}
}

// SubTestContextSetKV test setting KVs in context.
// Due to the incompatibility between gin and go-kit, the behavior is complicated,
// There are three "context" are relevant: func's ctx, http.Request.Context(), gin.Context.
// For Gin <1.8.0 + go-kit 0.11.0:
// 1. set KV via gin.Context anywhere would be reflected in all three contexts
// 2. set KV via web.SetKV on both func's ctx and http.Request.Context() anywhere would not be reflected on gin.Context
// 3. set KV to http.Request.Context() in middlewares would not be reflected on gin.Context
// 4. set KV to http.Request.Context() in endpoint would not take any effect after endpoint execution (thanks to go-kit...)
// For Gin >=1.8.0 + go-kit 0.11.0 with gin.Engine.ContextWithFallback enabled, all above rules are same except for
// - 2. would also reflacted in gin.Context
func SubTestContextSetKV(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// determine expected behaviors
		webKVSrc := kvSrcCtx | kvSrcReq
		if gVer := ginVersion(); gVer.major > 1 || gVer.minor >= 8 {
			logger.WithContext(ctx).Infof("Using updated rules for Gin v1.8.0+")
			webKVSrc = kvSrcAll
		}
		kvs := map[kv]kvSetter{
			makeKV("req", "req-value"):                           reqCtxKVSetter(),
			makeKV("gin-ctx", "gin-ctx-value", kvSrcAll):         ginCtxKVSetter(),
			makeKV("req-gin-ctx", "req-gin-ctx-value", kvSrcAll): reqGinCtxKVSetter(),
			makeKV("web-ctx", "web-ctx-value", webKVSrc):         webCtxKVSetter(),
			makeKV("web-req", "web-req-value", webKVSrc):         webReqKVSetter(),
		}

		overwrite := map[kv]kvSetter{
			makeKV("gin-ctx", "gin-ctx-new", kvSrcAll):         ginCtxKVSetter(),
			makeKV("req-gin-ctx", "req-gin-ctx-new", kvSrcAll): reqGinCtxKVSetter(),
			makeKV("web-ctx", "web-ctx-new", webKVSrc):         webCtxKVSetter(),
			makeKV("web-req", "web-req-new", webKVSrc):         webReqKVSetter(),
		}

		// setup middlewares and endpoints
		mwAssertion := func(ctx context.Context, req *http.Request) {
			for kv, setter := range kvs {
				setter(ctx, req, kv.k, kv.v)
			}
		}

		epAssertion := func(ctx context.Context, req *http.Request) {
			assertContextKVs(ctx, t, g, req, web.GinContext(ctx), "Endpoint", keys(kvs)...)
			for kv, setter := range overwrite {
				setter(ctx, req, kv.k, kv.v)
			}
		}

		ginAssertion := func(gc *gin.Context) {
			assertContextKVs(gc.Request.Context(), t, g, gc.Request, gc, "PreGinMW", keys(kvs)...)
			gc.Next()
			assertContextKVs(gc.Request.Context(), t, g, gc.Request, gc, "PostGinMW", keys(overwrite)...)
		}

		// execute test
		WebInit(ctx, t, g, di,
			registerAssertingEndpoint(http.MethodPost, "/mw/:var", epAssertion),
			registerAssertingMW(http.MethodPost, "/mw/**", mwAssertion),
			registerAssertingGinMW(http.MethodPost, "/mw/**", ginAssertion),
		)
		testEndpoint(ctx, t, g, http.MethodPost, "/mw/var-value")
	}
}

func SubTestGinHandlerMapping(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		handlerFn := func(gc *gin.Context) {
			g.Expect(web.GinContext(gc)).To(Not(BeNil()), "gin.Context from ctx should not be nil")
			g.Expect(web.GinContext(gc.Request.Context())).To(Not(BeNil()), "gin.Context from ctx should not be nil")
			g.Expect(web.HttpRequest(gc)).To(Equal(gc.Request), "web.HttpRequest should return same request")

			if resp, e := testdata.Raw(gc.Request.Context(), gc.Request); e != nil {
				gc.JSON(http.StatusInternalServerError, e)
			} else {
				gc.JSON(http.StatusOK, resp)
			}

		}

		ginMapping := web.NewSimpleGinMapping("gin", "/", "/mw/:var", http.MethodPost, nil, handlerFn)
		// execute test
		WebInit(ctx, t, g, di,
			func(reg *web.Registrar) {
				reg.MustRegister(ginMapping)
			},
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
			Order(0).
			Use(assertingMWFunc(fn)).
			Build())
	}
}

func registerAssertingGinMW(method, pattern string, fn gin.HandlerFunc) WebInitFunc {
	return func(reg *web.Registrar) {
		reg.MustRegister(middleware.NewBuilder("mw").
			ApplyTo(matcher.RouteWithPattern(pattern, method)).
			Order(1).
			Use(fn).
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

type ginVer struct {
	major int
	minor int
	patch int
}

var ginVerPattern = regexp.MustCompile(`^v(?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)`)

func ginVersion() ginVer {
	matches := ginVerPattern.FindSubmatch([]byte(gin.Version))
	if len(matches) < 4 {
		return ginVer{}
	}
	var ver ginVer
	if v, e := strconv.Atoi(string(matches[1])); e == nil {
		ver.major = v
	}
	if v, e := strconv.Atoi(string(matches[2])); e == nil {
		ver.minor = v
	}
	if v, e := strconv.Atoi(string(matches[3])); e == nil {
		ver.patch = v
	}
	return ver
}

type kvSrc int

const (
	kvSrcCtx kvSrc = 1 << iota
	kvSrcReq
	kvSrcGin
	kvSrcAll = kvSrcCtx | kvSrcReq | kvSrcGin
)

type kv struct {
	k   string
	v   string
	src kvSrc
}

func makeKV(k, v string, src ...kvSrc) kv {
	if len(src) == 0 {
		src = []kvSrc{kvSrcCtx, kvSrcReq}
	}
	var flag kvSrc
	for _, s := range src {
		flag |= s
	}
	return kv{k: k, v: v, src: flag}
}

func assertContextKVs(ctx context.Context, _ *testing.T, g *gomega.WithT, req *http.Request, gc *gin.Context, phase string, expectKVs ...kv) {
	for _, kv := range expectKVs {
		if kv.src&kvSrcCtx != 0 {
			g.Expect(ctx.Value(kv.k)).To(Equal(kv.v), "%s: context should contains correct %s=%s", phase, kv.k, kv.v)
		}
		if kv.src&kvSrcReq != 0 {
			g.Expect(req.Context().Value(kv.k)).To(Equal(kv.v), "%s: Request context should contains correct %s=%s", phase, kv.k, kv.v)
		}
		if kv.src&kvSrcGin != 0 {
			g.Expect(gc.Value(kv.k)).To(Equal(kv.v), "%s: Gin context should contain correct %s=%s", phase, kv.k, kv.v)
		}
	}
	g.Expect(ctx.Value("non-exist")).To(BeNil(), "%s: context should return nil on incorrect Key", phase)
	g.Expect(req.Context().Value("non-exist")).To(BeNil(), "%s: Request context should return nil on incorrect Key", phase)
	g.Expect(gc.Value("non-exist")).To(BeNil(), "%s: Gin context should return nil on incorrect Key", phase)
}

func keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k, _ := range m {
		keys = append(keys, k)
	}
	return keys
}

type kvSetter func(ctx context.Context, req *http.Request, k, v string)

func reqCtxKVSetter() kvSetter {
	return func(ctx context.Context, _ *http.Request, k, v string) {
		gc := web.GinContext(ctx)
		newCtx := context.WithValue(gc.Request.Context(), k, v)
		gc.Request = gc.Request.WithContext(newCtx)
	}
}

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
