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

package cors_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/cors"
    "github.com/cisco-open/go-lanai/pkg/web/rest"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"strings"
	"testing"
)

/*************************
	Setup Test
 *************************/

const (
	TestHeaderAllowed    = `X-Test-Header`
	TestHeaderExposed    = `X-Test-Header`
	TestHeaderDisallowed = `X-Test-Disallowed-Header`
)

func RegisterTestController(reg *web.Registrar) error {
	return reg.Register(TestController{})
}

/*************************
	Tests
 *************************/

func TestCORSGinMW(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(120*time.Second),
		webtest.WithMockedServer(),
		apptest.WithModules(cors.Module),
		apptest.WithFxOptions(
			fx.Invoke(RegisterTestController),
		),
		test.GomegaSubTest(SubTestPreFlightAllow(), "TestPreFlightAllow"),
		test.GomegaSubTest(SubTestPreFlightDisallowMethod(), "TestPreFlightDisallowMethod"),
		test.GomegaSubTest(SubTestPreFlightDisallowHeader(), "TestPreFlightDisallowHeader"),
		test.GomegaSubTest(SubTestPostSuccess(), "TestPostSuccess"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestPreFlightAllow() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        req := NewPreFlightRequest(ctx, http.MethodGet, "/hello", TestHeaderAllowed)
		resp := webtest.MustExec(ctx, req).Response
		g.Expect(resp).ToNot(BeNil(), "pre-flight response should not be nil")
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "status code should be correct")
        AssertHeaders(g, resp.Header, ExpectedPreFlightResponseHeader(http.MethodGet, TestHeaderAllowed), true)
	}
}

func SubTestPreFlightDisallowMethod() test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        req := NewPreFlightRequest(ctx, http.MethodDelete, "/hello", TestHeaderAllowed)
        resp := webtest.MustExec(ctx, req).Response
        g.Expect(resp).ToNot(BeNil(), "pre-flight response should not be nil")
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "status code should be correct")
        AssertHeaders(g, resp.Header, ExpectedPreFlightResponseHeader(http.MethodDelete, TestHeaderAllowed), false)
    }
}

func SubTestPreFlightDisallowHeader() test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        req := NewPreFlightRequest(ctx, http.MethodPost, "/hello", TestHeaderDisallowed)
        resp := webtest.MustExec(ctx, req).Response
        g.Expect(resp).ToNot(BeNil(), "pre-flight response should not be nil")
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "status code should be correct")
        AssertHeaders(g, resp.Header, ExpectedPreFlightResponseHeader(http.MethodPost, TestHeaderDisallowed), false)
    }
}

func SubTestPostSuccess() test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        req := NewCorsRequest(ctx, http.MethodPost, "/hello", TestHeaderAllowed, "test-request")
        resp := webtest.MustExec(ctx, req).Response
        g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "status code should be correct")
        AssertHeaders(g, resp.Header, ExpectedCorsResponseHeader(TestHeaderAllowed), true)
    }
}

/*************************
	Helpers
 *************************/

func NewPreFlightRequest(ctx context.Context, method, path string, headers ...string) *http.Request {
	return webtest.NewRequest(ctx, http.MethodOptions, path, nil, webtest.Headers(
		"Origin", "localhost",
		"Access-Control-Request-Method", method,
		"Access-Control-Request-Headers", strings.Join(headers, ", "),
	))
}

func NewCorsRequest(ctx context.Context, method, path string, headers ...string) *http.Request {
	headers = []string{"Origin", "localhost", "Authorization", "Bearer a_bearer_token"}
	return webtest.NewRequest(ctx, method, path, nil, webtest.Headers(headers...))
}


func ExpectedPreFlightResponseHeader(method string, headers...string) map[string]string {
    return map[string]string{
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Methods": method,
        "Access-Control-Allow-Headers": strings.Join(headers, ", "),
        "Access-Control-Allow-Credentials": "true",
        "Access-Control-Max-Age": "3600",
    }
}

func ExpectedCorsResponseHeader(headers...string) map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin": "*",
		"Access-Control-Expose-Headers": strings.Join(headers, ", "),
		"Access-Control-Allow-Credentials": "true",
		"X-Test-Header": "test-response",
	}
}

func HeaderAsMap(header http.Header) (m map[string]string) {
    m = make(map[string]string)
    for k := range header {
        m[k] = header.Get(k)
    }
    return
}

func AssertHeaders(g *gomega.WithT, header http.Header, expected map[string]string, expectExists bool) {
    h := HeaderAsMap(header)
    for k, v := range expected {
        if expectExists {
            g.Expect(h).To(HaveKeyWithValue(k, v), "response header should have kv [%s = %s]", k, v)
        } else {
            g.Expect(h).ToNot(HaveKey(k), "response header should not have key [%s]", k)
        }
    }
}

/*************************
	Dummy Controller
 *************************/

type TestController struct{}

func (c TestController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.Get("/hello").EndpointFunc(c.Hello).Build(),
		rest.Post("/hello").EndpointFunc(c.Hello).Build(),
		rest.Delete("/hello").EndpointFunc(c.Hello).Build(),
	}
}

func (TestController) Hello(_ context.Context, _ *http.Request) (*HeadererResponse, error) {
	resp := HeadererResponse{
		Header: make(http.Header),
		Msg:    "hello",
	}
	resp.Header.Set(TestHeaderExposed, "test-response")
	resp.Header.Set(TestHeaderDisallowed, "test-response")
	return &resp, nil
}

type HeadererResponse struct {
	Header http.Header `json:"-"`
	Msg    string      `json:"message"`
}

func (r *HeadererResponse) Headers() http.Header {
	return r.Header
}
