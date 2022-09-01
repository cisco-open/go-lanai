package examples

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"encoding/json"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"strings"
	"testing"
)

const (
	ValidRequestBody = `{"message": "hello"}`
)

/*************************
	Setup
 *************************/

type TestRequest struct {
	Message string `json:"message"`
}

type TestResponse struct {
	Message string `json:"message"`
}

type testController struct {}

func newTestController() web.Controller {
	return &testController{}
}

func (c *testController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.New("echo").Post("/api/v1/echo").
			EndpointFunc(c.Echo).Build(),
	}
}

func (c *testController) Echo(_ context.Context, req *TestRequest) (interface{}, error) {
	return &TestResponse{
		Message: req.Message,
	}, nil
}

/*************************
	Tests
 *************************/

type testDI struct {
	fx.In
}

func TestRealTestServer(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			web.FxControllerProviders(newTestController),
		),
		test.GomegaSubTest(SubTestEchoWithRelativePath(), "EchoWithRelativePath"),
		test.GomegaSubTest(SubTestEchoWithAbsolutePath(webtest.DefaultContextPath), "EchoWithAbsolutePath"),
	)
}

func TestMockedTestServer(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			web.FxControllerProviders(newTestController),
		),
		test.GomegaSubTest(SubTestEchoWithRelativePath(), "EchoWithRelativePath"),
		test.GomegaSubTest(SubTestEchoWithAbsolutePath(webtest.DefaultContextPath), "EchoWithAbsolutePath"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestEchoWithRelativePath() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response

		// with relative path
		req = webtest.NewRequest(ctx, http.MethodPost, "/api/v1/echo", strings.NewReader(ValidRequestBody),
			webtest.WithHeaders("Content-Type", "application/json"))
		resp = webtest.MustExec(ctx, req).Response
		assertResponse(t, g, resp, "hello")
	}
}

func SubTestEchoWithAbsolutePath(contextPath string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response

		// with absolute path
		url := fmt.Sprintf("http://whatever:0%s/api/v1/echo", contextPath)
		req = webtest.NewRequest(ctx, http.MethodPost, url, strings.NewReader(ValidRequestBody),
			webtest.WithHeaders("Content-Type", "application/json"))
		resp = webtest.MustExec(ctx, req).Response
		assertResponse(t, g, resp, "hello")
	}
}

/*************************
	Helpers
 *************************/

func assertResponse(_ *testing.T, g *gomega.WithT, resp *http.Response, expectedMsg string) {
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "response should be 200")
	var tsBody TestResponse
	e := json.NewDecoder(resp.Body).Decode(&tsBody)
	g.Expect(e).To(Succeed(), "parsing response body shouldn't fail")
	g.Expect(tsBody.Message).To(Equal(expectedMsg), "response body should be correct")
}

