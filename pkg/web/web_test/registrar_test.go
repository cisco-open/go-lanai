package web_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/web_test/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
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

func TestRestRegistration(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithUtilities(),
		//apptest.WithModules(),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(web.NewEngine),
		),
		test.SubTestSetup(ResetEngine(&di)),
		test.GomegaSubTest(SubTestWithController(&di), "TestWithController"),
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

		resp := testEndpoint(ctx, t, g, http.MethodPost, "/basic/var-value")
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "response status code should be correct")
		body, e := io.ReadAll(resp.Body)
		g.Expect(e).To(Succeed(), "reading response body should success")
		g.Expect(body).To(Not(BeEmpty()), "response body should not be empty")
		//g.Expect(body).To(HaveJson(http.StatusOK), "response status code should be correct")
	}
}

/*************************
	Helpers
 *************************/

func testEndpoint(ctx context.Context, _ *testing.T, g *gomega.WithT, method, path string, opts ...webtest.RequestOptions) *http.Response {
	basicOpts := []webtest.RequestOptions{
		webtest.Headers(BasicHeaderKey, BasicHeaderValue),
		webtest.Queries(BasicQueryKey, BasicQueryValue),
	}

	req := webtest.NewRequest(ctx, method, path, strings.NewReader(BasicBody), append(basicOpts, opts...)...)
	resp := webtest.MustExec(ctx, req).Response
	g.Expect(resp).To(Not(BeNil()), "response should not be nil")
	return resp
}
