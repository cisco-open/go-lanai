package web_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/web_test/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

/*************************
	Setup
 *************************/

const (
	BasicBody        = `{"string":"string value","int":20}`
	BasicHeaderKey   = `X-VAR`
	BasicHeaderValue = `header-value`
	BasicQueryKey    = `q`
	BasicQueryValue  = `query-value`
)

type TestDI struct {
	fx.In
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
	return web.NewRegistrar(di.Engine, di.Properties)
}

/*************************
	Tests
 *************************/

func TestMvcRegistration(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithUtilities(),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(web.NewEngine),
		),
		test.SubTestSetup(ResetEngine(&di)),
		test.GomegaSubTest(SubTestWithController(&di), "TestWithController"),
		test.GomegaSubTest(SubTestWithMvcMapping(&di), "TestWithMvcMapping"),
		test.GomegaSubTest(SubTestWithMvcVariations(&di), "TestWithMvcVariations"),
		test.GomegaSubTest(SubTestWithInvalidMvcHandler(&di), "TestWithInvalidMvcHandler"),
		test.GomegaSubTest(SubTestWithTextResponse(&di), "TestWithTextResponse"),
		test.GomegaSubTest(SubTestWithBytesResponse(&di), "TestWithBytesResponse"),
		test.GomegaSubTest(SubTestWithCustomResponseEncoder(&di), "TestWithCustomResponseEncoder"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestWithController(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		reg := NewTestRegister(di)
		var e error
		reg.MustRegister(web.NewLoggingCustomizer(di.Properties))
		e = reg.Register(testdata.Controller{})
		g.Expect(e).To(Succeed(), "register controller should success")

		e = reg.Initialize(ctx)
		g.Expect(e).To(Succeed(), "initialize should success")
		testBasicEndpoint(ctx, t, g, http.MethodPost, "/basic/var-value")
	}
}

func SubTestWithMvcMapping(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		reg := NewTestRegister(di)
		var e error
		reg.MustRegister(web.NewLoggingCustomizer(di.Properties))
		e = reg.Register(rest.Post("/basic/:var").EndpointFunc(testdata.StructPtr200).Build())
		g.Expect(e).To(Succeed(), "register MVC mapping should success")

		e = reg.Initialize(ctx)
		g.Expect(e).To(Succeed(), "initialize should success")
		testBasicEndpoint(ctx, t, g, http.MethodPost, "/basic/var-value")
	}
}

func SubTestWithMvcVariations(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const uriVar = `:var`
		variations := map[web.MvcMapping]func(*expectation){
			rest.Post("/struct/ptr/" + uriVar).EndpointFunc(testdata.StructPtr200).Build(): nil,
			rest.Post("/struct/" + uriVar).EndpointFunc(testdata.Struct200).Build():        nil,
			rest.Post("/struct/ptr/201/" + uriVar).EndpointFunc(testdata.StructPtr201).Build(): func(expect *expectation) {
				expect.status = http.StatusCreated
			},
			rest.Post("/struct/201/" + uriVar).EndpointFunc(testdata.Struct201).Build(): func(expect *expectation) {
				expect.status = http.StatusCreated
			},
			rest.Post("/struct/ptr/201/header" + uriVar).EndpointFunc(testdata.StructPtr201WithHeader).Build(): func(expect *expectation) {
				expect.status = http.StatusCreated
				expect.headers = map[string]string{BasicHeaderKey: BasicHeaderValue}
			},
			rest.Post("/struct/201/header" + uriVar).EndpointFunc(testdata.Struct201WithHeader).Build(): func(expect *expectation) {
				expect.status = http.StatusCreated
				expect.headers = map[string]string{BasicHeaderKey: BasicHeaderValue}
			},
			rest.Post("/raw/" + uriVar).EndpointFunc(testdata.Raw).Build(): nil,
			rest.Post("/no/request/" + uriVar).EndpointFunc(testdata.NoRequest).Build(): func(expect *expectation) {
				expect.body = map[string]interface{}{"uri": "", "q": "", "header": "", "string": "", "int": float64(0)}
			},
		}

		reg := NewTestRegister(di)
		var e error
		reg.MustRegister(web.NewLoggingCustomizer(di.Properties))

		// test registration
		mappings := make([]web.MvcMapping, 0, len(variations))
		for k := range variations {
			mappings = append(mappings, k)
		}
		e = reg.Register(mappings)
		g.Expect(e).To(Succeed(), "register MVC mappings should success")

		e = reg.Initialize(ctx)
		g.Expect(e).To(Succeed(), "initialize should success")
		for k, v := range variations {
			method := k.Method()
			path := strings.ReplaceAll(k.Path(), uriVar, "var-value")
			testBasicEndpoint(ctx, t, g, method, path, v)
		}
	}
}

func SubTestWithInvalidMvcHandler(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {

		variations := []rest.EndpointFunc{
			testdata.MissingResponse,
			testdata.MissingError,
			testdata.MissingContext,
			testdata.WrongErrorPosition,
			testdata.WrongContextPosition,
			testdata.ExtraInput,
			"Not a Func",
		}

		reg := NewTestRegister(di)
		var e error
		reg.MustRegister(web.NewLoggingCustomizer(di.Properties))

		// test registration
		for _, v := range variations {
			e = withRecover(func() error {
				return reg.Register(rest.Post("/invalid").EndpointFunc(v).Build())
			})
			g.Expect(e).To(HaveOccurred(), "register MVC mapping [%T] should fail", v)
		}
	}
}

func SubTestWithTextResponse(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		reg := NewTestRegister(di)
		var e error
		reg.MustRegister(web.NewLoggingCustomizer(di.Properties))
		mappings := []web.MvcMapping {
			rest.Post("/text/:var").EndpointFunc(testdata.Text).EncodeResponseFunc(web.TextResponseEncoder()).Build(),
			rest.Post("/text/string/:var").EndpointFunc(testdata.TextString).EncodeResponseFunc(web.TextResponseEncoder()).Build(),
			rest.Post("/text/bytes/:var").EndpointFunc(testdata.TextBytes).EncodeResponseFunc(web.TextResponseEncoder()).Build(),
		}
		e = reg.Register(mappings)
		g.Expect(e).To(Succeed(), "register MVC mapping should success")

		e = reg.Initialize(ctx)
		g.Expect(e).To(Succeed(), "initialize should success")
		for _, m := range mappings {
			path := strings.ReplaceAll(m.Path(), ":var", "var-value")
			testBasicEndpoint(ctx, t, g, m.Method(), path, func(expect *expectation) {
				expect.headers = map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				}
				expect.bodyDecoder = urlencodedBodyDecoder()
				expect.body["int"] = "20"
			})
		}
	}
}

func SubTestWithBytesResponse(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		reg := NewTestRegister(di)
		var e error
		reg.MustRegister(web.NewLoggingCustomizer(di.Properties))
		mappings := []web.MvcMapping {
			rest.Post("/bytes/:var").EndpointFunc(testdata.Bytes).EncodeResponseFunc(web.BytesResponseEncoder()).Build(),
			rest.Post("/bytes/string/:var").EndpointFunc(testdata.BytesString).EncodeResponseFunc(web.BytesResponseEncoder()).Build(),
			rest.Post("/bytes/struct/:var").EndpointFunc(testdata.BytesStruct).EncodeResponseFunc(web.BytesResponseEncoder()).Build(),
		}
		e = reg.Register(mappings)
		g.Expect(e).To(Succeed(), "register MVC mapping should success")

		e = reg.Initialize(ctx)
		g.Expect(e).To(Succeed(), "initialize should success")
		for _, m := range mappings {
			path := strings.ReplaceAll(m.Path(), ":var", "var-value")
			testBasicEndpoint(ctx, t, g, m.Method(), path, func(expect *expectation) {
				expect.headers = map[string]string{
					"Content-Type": "application/octet-stream",
				}
				expect.bodyDecoder = urlencodedBodyDecoder()
				expect.body["int"] = "20"
			})
		}
	}
}

func SubTestWithCustomResponseEncoder(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const customContentType = `application/also-json`
		reg := NewTestRegister(di)
		var e error
		reg.MustRegister(web.NewLoggingCustomizer(di.Properties))
		encoder := web.CustomResponseEncoder(func(opt *web.EncodeOption) {
			opt.ContentType = customContentType
			opt.WriteFunc = web.JsonWriteFunc
		})
		e = reg.Register(rest.Post("/basic/:var").EndpointFunc(testdata.StructPtr200).EncodeResponseFunc(encoder).Build())
		g.Expect(e).To(Succeed(), "register MVC mapping should success")

		e = reg.Initialize(ctx)
		g.Expect(e).To(Succeed(), "initialize should success")
		testBasicEndpoint(ctx, t, g, http.MethodPost, "/basic/var-value", func(expect *expectation) {
			expect.headers = map[string]string{
				"Content-Type": customContentType,
			}
		})
	}
}

/*************************
	Helpers
 *************************/

func testBasicEndpoint(ctx context.Context, t *testing.T, g *gomega.WithT, method, path string, expects ...func(expect *expectation)) {
	resp := invokeEndpoint(ctx, t, g, method, path)
	expect := expectation{
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

type expectation struct {
	status  int
	headers map[string]string
	body    map[string]interface{}
	bodyDecoder bodyDecoder
}

type bodyDecoder func(body io.Reader) (interface{}, error)

func assertResponse(_ *testing.T, g *gomega.WithT, resp *http.Response, expect expectation) {
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

func jsonBodyDecoder() bodyDecoder {
	return func(body io.Reader) (interface{}, error) {
		decoder := json.NewDecoder(body)
		var i interface{}
		if e := decoder.Decode(&i); e != nil {
			return nil, e
		}
		return i, nil
	}
}

func urlencodedBodyDecoder() bodyDecoder {
	return func(body io.Reader) (interface{}, error) {
		text, e := io.ReadAll(body)
		if e != nil {
			return nil, e
		}
		pairs := strings.Split(string(text), "&")
		ret := map[string]interface{}{}
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) != 2 {
				continue
			}
			v, e := url.QueryUnescape(kv[1])
			if e != nil {
				continue
			}
			ret[kv[0]] = v
		}
		return ret, nil
	}
}

func withRecover(fn func() error) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("%v", v)
		}
	}()
	return fn()
}
