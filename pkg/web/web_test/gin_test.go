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
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/log"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/matcher"
    "github.com/cisco-open/go-lanai/pkg/web/middleware"
    "github.com/cisco-open/go-lanai/pkg/web/rest"
    "github.com/cisco-open/go-lanai/pkg/web/web_test/testdata"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/webtest"
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
		webtest.WithUtilities(webtest.UseContextPath("/custom-prefix")),
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
		assertion := func(ctx context.Context, req *http.Request) {
			g.Expect(web.ContextPath(ctx)).To(Equal(webtest.CurrentContextPath(ctx)), "web.ContextPath should return correct value")
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
		const (
			kWebCtx = 0xff
			kWebReq = 3.14
		)
		webKVSrc := kvSrcCtx | kvSrcReq
		if gVer := ginVersion(); gVer.major > 1 || gVer.minor >= 8 {
			logger.WithContext(ctx).Infof("Using updated rules for Gin v1.8.0+")
			webKVSrc = kvSrcAll
		}
		kvs := []kv{
			makeKV("req", "req-value", reqCtxKVSetter()),
			makeKV("gin-ctx", "gin-ctx-value", ginCtxKVSetter(), kvSrcAll),
			makeKV("req-gin-ctx", "req-gin-ctx-value", reqGinCtxKVSetter(), kvSrcAll),
			makeKV(kWebCtx, "web-ctx-value", webCtxKVSetter(), webKVSrc),
			makeKV(kWebReq, "web-req-value", webReqKVSetter(), webKVSrc),
		}

		overwrite := []kv{
			makeKV("gin-ctx", "gin-ctx-new", ginCtxKVSetter(), kvSrcAll),
			makeKV("req-gin-ctx", "req-gin-ctx-new", reqGinCtxKVSetter(), kvSrcAll),
			makeKV("web-ctx", "web-ctx-new",  webCtxKVSetter(), webKVSrc),
			makeKV("web-req", "web-req-new", webReqKVSetter(), webKVSrc),
		}

		// setup middlewares and endpoints
		mwAssertion := func(ctx context.Context, req *http.Request) {
			for _, kv := range kvs {
				kv.setter(ctx, req, kv.k, kv.v)
			}
		}

		epAssertion := func(ctx context.Context, req *http.Request) {
			assertContextKVs(ctx, t, g, req, web.GinContext(ctx), "Endpoint", kvs...)
			for _, kv := range overwrite {
				kv.setter(ctx, req, kv.k, kv.v)
			}
		}

		ginAssertion := func(gc *gin.Context) {
			assertContextKVs(gc.Request.Context(), t, g, gc.Request, gc, "PreGinMW", kvs...)
			gc.Next()
			assertContextKVs(gc.Request.Context(), t, g, gc.Request, gc, "PostGinMW", overwrite...)
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

type kvSetter func(ctx context.Context, req *http.Request, k, v any)

type kv struct {
	k      interface{}
	v      interface{}
	setter kvSetter
	src    kvSrc
}

func makeKV(k, v any, setter kvSetter, src ...kvSrc) kv {
	if len(src) == 0 {
		src = []kvSrc{kvSrcCtx, kvSrcReq}
	}
	var flag kvSrc
	for _, s := range src {
		flag |= s
	}
	return kv{k: k, v: v, setter: setter, src: flag}
}

func assertContextKVs(ctx context.Context, _ *testing.T, g *gomega.WithT, req *http.Request, gc *gin.Context, phase string, expectKVs ...kv) {

	for _, kv := range expectKVs {
		if kv.src&kvSrcCtx != 0 {
			g.Expect(ctx.Value(kv.k)).To(Equal(kv.v), "%s: context should contains correct %v=%v", phase, kv.k, kv.v)
		}
		if kv.src&kvSrcReq != 0 {
			g.Expect(req.Context().Value(kv.k)).To(Equal(kv.v), "%s: Request context should contains correct %v=%v", phase, kv.k, kv.v)
		}
		if kv.src&kvSrcGin != 0 {
			g.Expect(gc.Value(kv.k)).To(Equal(kv.v), "%s: Gin context should contain correct %v=%v", phase, kv.k, kv.v)
		}
	}
	g.Expect(ctx.Value("non-exist")).To(BeNil(), "%s: context should return nil on incorrect Key", phase)
	g.Expect(req.Context().Value("non-exist")).To(BeNil(), "%s: Request context should return nil on incorrect Key", phase)
	g.Expect(gc.Value("non-exist")).To(BeNil(), "%s: Gin context should return nil on incorrect Key", phase)
}

func reqCtxKVSetter() kvSetter {
	return func(ctx context.Context, _ *http.Request, k, v interface{}) {
		gc := web.GinContext(ctx)
		newCtx := context.WithValue(gc.Request.Context(), k, v)
		gc.Request = gc.Request.WithContext(newCtx)
	}
}

func ginCtxKVSetter() kvSetter {
	return func(ctx context.Context, req *http.Request, k, v interface{}) {
		gc := web.GinContext(ctx)
		gc.Set(fmt.Sprintf(`%v`, k), v)
	}
}

func reqGinCtxKVSetter() kvSetter {
	return func(ctx context.Context, req *http.Request, k, v interface{}) {
		gc := web.GinContext(req.Context())
		gc.Set(fmt.Sprintf(`%v`, k), v)
	}
}

func webCtxKVSetter() kvSetter {
	return func(ctx context.Context, req *http.Request, k, v interface{}) {
		web.SetKV(ctx, k, v)
	}
}

func webReqKVSetter() kvSetter {
	return func(ctx context.Context, req *http.Request, k, v interface{}) {
		web.SetKV(req.Context(), k, v)
	}
}
