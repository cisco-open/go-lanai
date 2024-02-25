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

package webtest

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/log"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/rest"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "net/http"
    "strings"
    "testing"
)

const (
	ValidRequestBody = `{"message": "hello"}`
	ValidQuery       = `hello`
)

/*************************
	Setup
 *************************/

type TestRequest struct {
	Message string `json:"message"`
	Query   string `form:"q"`
}

type TestResponse struct {
	Message string `json:"message"`
	Query   string `json:"query"`
}

type testController struct{}

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
		Query:   req.Query,
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
		test.GomegaSubTest(SubTestEchoWithRelativePath(false), "EchoWithRelativePath"),
		test.GomegaSubTest(SubTestEchoWithAbsolutePath(DefaultContextPath, false), "EchoWithAbsolutePath"),
	)
}

func TestCustomRealTestServer(t *testing.T) {
	const altContextPath = "/also-test"
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithRealServer(
			UseContextPath(altContextPath),
			UsePort(0),
			UseLogLevel(log.LevelDebug),
			AddDefaultRequestOptions(Queries("q", ValidQuery)),
		),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			web.FxControllerProviders(newTestController),
		),
		test.GomegaSubTest(SubTestRealServerUtils(0, altContextPath), "TestRealServerUtils"),
		test.GomegaSubTest(SubTestEchoWithRelativePath(true), "EchoWithRelativePath"),
		test.GomegaSubTest(SubTestEchoWithAbsolutePath(altContextPath, true), "EchoWithAbsolutePath"),
	)
}

func TestDefaultMockedTestServer(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithMockedServer(),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			web.FxControllerProviders(newTestController),
		),
		test.GomegaSubTest(SubTestEchoWithRelativePath(false), "EchoWithRelativePath"),
		test.GomegaSubTest(SubTestEchoWithAbsolutePath(DefaultContextPath, false), "EchoWithAbsolutePath"),
	)
}

func TestCustomMockedTestServer(t *testing.T) {
	const altContextPath = "/also-test"
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithMockedServer(
			UseContextPath(altContextPath),
			UseLogLevel(log.LevelDebug),
			AddDefaultRequestOptions(Queries("q", ValidQuery)),
		),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			web.FxControllerProviders(newTestController),
		),
		test.GomegaSubTest(SubTestEchoWithRelativePath(true), "EchoWithRelativePath"),
		test.GomegaSubTest(SubTestEchoWithAbsolutePath(altContextPath, true), "EchoWithAbsolutePath"),
	)
}

func TestUtilities(t *testing.T) {
	const altContextPath = "/also-test"
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithUtilities(
			UseContextPath(altContextPath),
			UseLogLevel(log.LevelDebug),
			AddDefaultRequestOptions(Queries("q", ValidQuery)),
		),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(
				web.NewEngine,
				web.NewRegistrar,
			),
			fx.Invoke(initialize),
			web.FxControllerProviders(newTestController),
		),
		test.GomegaSubTest(SubTestEchoWithRelativePath(true), "EchoWithRelativePath"),
		test.GomegaSubTest(SubTestEchoWithAbsolutePath(altContextPath, true), "EchoWithAbsolutePath"),
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

func SubTestEchoWithRelativePath(expectQuery bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response

		// with relative path
		req = NewRequest(ctx, http.MethodPost, "/api/v1/echo", strings.NewReader(ValidRequestBody))
		req.Header.Set("Content-Type", "application/json")
		resp = MustExec(ctx, req).Response
		assertResponse(t, g, resp, "hello", expectQuery)
	}
}

func SubTestEchoWithAbsolutePath(contextPath string, expectQuery bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response

		// with absolute path
		url := fmt.Sprintf("http://whatever:0%s/api/v1/echo", contextPath)
		req = NewRequest(ctx, http.MethodPost, url, strings.NewReader(ValidRequestBody))
		req.Header.Set("Content-Type", "application/json")
		resp = MustExec(ctx, req).Response
		assertResponse(t, g, resp, "hello", expectQuery)
	}
}

/*************************
	Helpers
 *************************/

func assertResponse(_ *testing.T, g *gomega.WithT, resp *http.Response, expectedMsg string, expectQuery bool) {
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "response should be 200")
	var tsBody TestResponse
	e := json.NewDecoder(resp.Body).Decode(&tsBody)
	g.Expect(e).To(Succeed(), "parsing response body shouldn't fail")
	g.Expect(tsBody.Message).To(Equal(expectedMsg), "response body should have correct message")
	if expectQuery {
		g.Expect(tsBody.Query).To(Equal(ValidQuery), "response body should have correct query")
	}
}
