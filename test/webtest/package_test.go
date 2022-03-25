package webtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
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

func TestDefaultRealTestServer(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithRealServer(),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			web.FxControllerProviders(newTestController),
		),
		test.GomegaSubTest(SubTestRealServerUtils(0, DefaultContextPath), "TestRealServerUtils"),
		test.GomegaSubTest(SubTestEchoWithRelativePath(), "EchoWithRelativePath"),
		test.GomegaSubTest(SubTestEchoWithAbsolutePath(DefaultContextPath), "EchoWithAbsolutePath"),
	)
}

func TestCustomRealTestServer(t *testing.T) {
	const altContextPath = "/also-test"
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithRealServer(UseContextPath(altContextPath), UseLogLevel(log.LevelDebug)),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			web.FxControllerProviders(newTestController),
		),
		test.GomegaSubTest(SubTestRealServerUtils(0, altContextPath), "TestRealServerUtils"),
		test.GomegaSubTest(SubTestEchoWithRelativePath(), "EchoWithRelativePath"),
		test.GomegaSubTest(SubTestEchoWithAbsolutePath(altContextPath), "EchoWithAbsolutePath"),
	)
}

func TestMockedTestServer(t *testing.T) {
	const altContextPath = "/also-test"
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithMockedServer(UseContextPath(altContextPath), UseLogLevel(log.LevelDebug)),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			web.FxControllerProviders(newTestController),
		),
		test.GomegaSubTest(SubTestEchoWithRelativePath(), "EchoWithRelativePath"),
		test.GomegaSubTest(SubTestEchoWithAbsolutePath(altContextPath), "EchoWithAbsolutePath"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestRealServerUtils(expectedPort int, expectedContextPath string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		port := CurrentPort(ctx)
		if expectedPort <= 0 {
			g.Expect(port).To(BeNumerically(">", 0), "CurrentPort should return valid value")
		} else {
			g.Expect(port).To(BeNumerically("==", expectedPort), "CurrentPort should return correct value")
		}

		ctxPath := CurrentContextPath(ctx)
		g.Expect(ctxPath).To(Equal(expectedContextPath), "CurrentContextPath should returns correct path")
	}
}

func SubTestEchoWithRelativePath() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response

		// with relative path
		req = NewRequest(ctx, http.MethodPost, "/api/v1/echo", strings.NewReader(ValidRequestBody))
		req.Header.Set("Content-Type", "application/json")
		resp = MustExec(ctx, req).Response
		assertResponse(t, g, resp, "hello")
	}
}

func SubTestEchoWithAbsolutePath(contextPath string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response

		// with absolute path
		url := fmt.Sprintf("http://whatever:0%s/api/v1/echo", contextPath)
		req = NewRequest(ctx, http.MethodPost, url, strings.NewReader(ValidRequestBody))
		req.Header.Set("Content-Type", "application/json")
		resp = MustExec(ctx, req).Response
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

