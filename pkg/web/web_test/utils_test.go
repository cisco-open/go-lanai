package web_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"strings"
	"testing"
)

/*************************
	Common Setup
 *************************/

const (
	BasicBody        = `{"string":"string value","int":20}`
	BasicHeaderKey   = `X-VAR`
	BasicHeaderValue = `header-value`
	BasicQueryKey    = `q`
	BasicQueryValue  = `query-value`
)

type TestDI struct {
	fx.In      `ignore-unexported:"true"`
	Engine     *web.Engine
	Properties web.ServerProperties
}

// ResetRegister reset gin engine to a clean state
func ResetEngine(di *TestDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		di.Engine.Engine = gin.New()
		return ctx, nil
	}
}

func NewTestRegister(di *TestDI) *web.Registrar {
	reg := web.NewRegistrar(di.Engine, di.Properties)
	return reg
}

type WebInitFunc func(reg *web.Registrar)

func WebInit(ctx context.Context, _ *testing.T, g *gomega.WithT, di *TestDI, initFn ...WebInitFunc) {
	reg := NewTestRegister(di)
	reg.MustRegister(web.NewLoggingCustomizer(di.Properties))
	for _, fn := range initFn {
		fn(reg)
	}
	e := reg.Initialize(ctx)
	g.Expect(e).To(Succeed(), "initialize should success")
}

/*************************
	MVC
 *************************/

func testEndpoint(ctx context.Context, t *testing.T, g *gomega.WithT, method, path string, expects ...func(expect *mvcExpectation)) {
	resp := invokeEndpoint(ctx, t, g, method, path)
	expect := mvcExpectation{
		status: http.StatusOK,
		headers: map[string]string{
			"Content-Type": "application/json; charset=utf-8",
		},
		body: map[string]interface{}{
			"uri":    "var-value",
			"q":      BasicQueryValue,
			"header": BasicHeaderValue,
			"string": "string value",
			"int":    float64(20),
		},
		bodyDecoder: jsonBodyDecoder(),
	}
	for _, fn := range expects {
		if fn != nil {
			fn(&expect)
		}
	}
	assertResponse(t, g, resp, expect)
}

func invokeEndpoint(ctx context.Context, _ *testing.T, g *gomega.WithT, method, path string, opts ...webtest.RequestOptions) *http.Response {
	basicOpts := []webtest.RequestOptions{
		webtest.Headers("Content-Type", "application/json", BasicHeaderKey, BasicHeaderValue),
		webtest.Queries(BasicQueryKey, BasicQueryValue),
	}

	req := webtest.NewRequest(ctx, method, path, strings.NewReader(BasicBody), append(basicOpts, opts...)...)
	resp := webtest.MustExec(ctx, req).Response
	g.Expect(resp).To(Not(BeNil()), "response should not be nil")
	return resp
}

type mvcExpectation struct {
	status      int
	headers     map[string]string
	body        map[string]interface{}
	bodyDecoder bodyDecoder
}

func assertResponse(_ *testing.T, g *gomega.WithT, resp *http.Response, expect mvcExpectation) {
	g.Expect(resp.StatusCode).To(Equal(expect.status), "response status code should be correct")
	for k, v := range expect.headers {
		g.Expect(resp.Header.Get(k)).To(Equal(v), "response header should have header %s", k)
	}

	if expect.body != nil && expect.bodyDecoder != nil {
		body, e := expect.bodyDecoder(resp.Body)
		g.Expect(e).To(Succeed(), "decode response body should success")
		g.Expect(body).To(BeEquivalentTo(expect.body), "response body should be correct")
	}
}

/*************************
	Middleware
 *************************/

type mwInvocation struct {
	gc *gin.Context
	rw http.ResponseWriter
	r  *http.Request
}

type TestMW struct {
	Invocation []mwInvocation
}

func NewTestMW() *TestMW {
	return &TestMW{}
}

func (mw *TestMW) Invoked() int {
	return len(mw.Invocation)
}

func (mw *TestMW) Reset() {
	mw.Invocation = nil
}

func (mw *TestMW) HandlerFunc() web.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		mw.Invocation = append(mw.Invocation, mwInvocation{rw: rw, r:  r, gc: web.GinContext(r.Context())})
	}
}

func (mw *TestMW) HttpHandlerFunc() http.HandlerFunc {
	return http.HandlerFunc(mw.HandlerFunc())
}

func (mw *TestMW) GinHandlerFunc() gin.HandlerFunc {
	return func(gc *gin.Context) {
		mw.Invocation = append(mw.Invocation, mwInvocation{
			gc: gc,
			rw: gc.Writer,
			r:  gc.Request,
		})
	}
}

type mwExpectation struct {
	count int
	single bool
}

func mwExpectCount(invocationCount int) func(expect *mwExpectation) {
	return func(expect *mwExpectation) {
		expect.count = invocationCount
	}
}

func assertMW(_ *testing.T, g *gomega.WithT, mw *TestMW, expects ...func(expect *mwExpectation)) {
	expect := mwExpectation{
		count: 1,
		single: true,
	}
	for _, fn := range expects {
		if fn != nil {
			fn(&expect)
		}
	}
	g.Expect(mw.Invocation).To(HaveLen(expect.count), "Middleware should be invoked correctly")
	var prev *mwInvocation
	for i, v := range mw.Invocation {
		g.Expect(v.r).To(Not(BeNil()), "invocation's request should not be nil")
		g.Expect(v.rw).To(Not(BeNil()), "invocation's response writer should not be nil")
		g.Expect(v.gc).To(Not(BeNil()), "invocation's gin.Context should not be nil")
		if expect.single && prev != nil {
			g.Expect(v.r).To(Equal(prev.r), "invocation's request should be same as previous invocation")
			g.Expect(v.rw).To(Equal(prev.rw), "invocation's response writer should be same as previous invocation")
			g.Expect(v.gc).To(Equal(prev.gc), "invocation's gin.Context should be same as previous invocation")
		}
		prev = &mw.Invocation[i]
	}
}