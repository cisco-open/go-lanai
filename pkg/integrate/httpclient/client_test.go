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

package httpclient_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/integrate/httpclient"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/sdtest"
	gomegautils "github.com/cisco-open/go-lanai/test/utils/gomega"
	"github.com/cisco-open/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"net/url"
	"path"
	"testing"
	"time"
)

/*************************
	Setup
 *************************/

const (
	SDServiceNameFullInfo  = `mockedserver`
	SDServiceNamePortOnly  = `mockedserver-port-only`
	SDServiceNameNoInfo    = `mockedserver-no-info`
	TestPath               = "/echo"
	TestErrorPath          = "/fail"
	TestNoContentPath      = "/nocontent"
	TestNoContentErrorPath = "/nocontentfail"
	TestMaybeFailPath      = "/maybe"
	TestTimeoutPath        = "/timeout"
)

// UpdateMockedSD update SD record to use the random server port
func UpdateMockedSD(di *TestDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		port := webtest.CurrentPort(ctx)
		if port <= 0 {
			return ctx, nil
		}
		di.Client.UpdateMockedService(SDServiceNameFullInfo, sdtest.NthInstance(0), func(inst *discovery.Instance) {
			inst.Port = port
		})
		di.Client.UpdateMockedService(SDServiceNamePortOnly, sdtest.NthInstance(0), func(inst *discovery.Instance) {
			inst.Port = port
		})
		di.Client.UpdateMockedService(SDServiceNameNoInfo, sdtest.NthInstance(0), func(inst *discovery.Instance) {
			inst.Port = 0
		})
		return ctx, nil
	}
}

/*************************
	Tests
 *************************/

type TestDI struct {
	fx.In
	sdtest.DI
	HttpClient       httpclient.Client
	MockedController *MockedController
}

func TestWithMockedServer(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		sdtest.WithMockedSD(sdtest.DefinitionWithPrefix("mocks.sd")),
		apptest.WithModules(httpclient.Module),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(NewMockedController),
			web.FxControllerProviders(ProvideWebController),
		),
		test.SubTestSetup(UpdateMockedSD(&di)),
		test.GomegaSubTest(SubTestWithFullInfoSD(&di), "TestWithFullInfoSD"),
		test.GomegaSubTest(SubTestWithPortOnlySD(&di), "TestWithPortOnlySD"),
		test.GomegaSubTest(SubTestWithNoInfoSD(&di), "TestWithNoInfoSD"),
		test.GomegaSubTest(SubTestWithSDNoResponseContent(&di), "TestWithSDNoResponseContent"),
		test.GomegaSubTest(SubTestWithBaseURL(&di), "TestWithBaseURL"),
		test.GomegaSubTest(SubTestWithBaseURLFailure(&di), "TestWithBaseUrlFailure"),
		test.GomegaSubTest(SubTestWithSDFailure(&di), "TestWithSDFailure"),
		test.GomegaSubTest(SubTestWithErrorResponse(&di), "TestWithErrorResponse"),
		test.GomegaSubTest(SubTestWithUnexpectedStatusCodeInErrorResponse(&di), "TestWithUnexpectedStatusCodeInErrorResponse"),
		test.GomegaSubTest(SubTestWithNoContentErrorResponse(&di), "SubTestWithNoContentErrorResponse"),
		test.GomegaSubTest(SubTestWithFailedSD(&di), "TestWithFailedSD"),
		test.GomegaSubTest(SubTestWithRetry(&di), "TestWithRetry"),
		test.GomegaSubTest(SubTestWithTimeout(&di), "TestWithTimeout"),
		test.GomegaSubTest(SubTestWithURLEncoded(&di), "TestWithURLEncoded"),
		test.GomegaSubTest(SubTestWithAbsoluteUrl(&di), "TestWithAbsoluteUrl"),
	)
}

/*************************
	Sub Tests
 *************************/

// SubTestWithFullInfoSD discovered service has information about port, scheme and context-path (from meta or tags)
func SubTestWithFullInfoSD(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithService(SDServiceNameFullInfo)
		g.Expect(e).To(Succeed(), "client with service name should be available")
		performEchoTest(ctx, t, g, client)
	}
}

// SubTestWithPortOnlySD discovered service has no information about scheme and context-path, only has "port
func SubTestWithPortOnlySD(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var client httpclient.Client
		var e error
		// no extra SD options, should fail
		client, e = di.HttpClient.WithService(SDServiceNamePortOnly)
		g.Expect(e).To(Succeed(), "client with service name should be available")
		reqBody := makeEchoRequestBody()
		req := httpclient.NewRequest(TestPath, http.MethodPost, httpclient.WithBody(reqBody))
		_, e = client.Execute(ctx, req, httpclient.JsonBody(&EchoResponse{}))
		g.Expect(e).To(HaveOccurred(), "execution should fail without extra SD options")
		g.Expect(e).To(gomegautils.IsError(httpclient.ErrorTypeInternal), "error should be correct without extra SD options")

		// with proper SD options, should not fail
		client, e = di.HttpClient.WithService(SDServiceNamePortOnly, func(opt *httpclient.SDOption) {
			opt.Scheme = "http"
			opt.ContextPath = "/test"
		})
		g.Expect(e).To(Succeed(), "client with service name should be available")
		performEchoTest(ctx, t, g, client)
	}
}

// SubTestWithNoInfoSD discovered service only have address/IP, no port, scheme nor context-path information
func SubTestWithNoInfoSD(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var client httpclient.Client
		var e error
		client, e = di.HttpClient.WithService(SDServiceNameNoInfo, func(opt *httpclient.SDOption) {
			opt.Scheme = "http"
			opt.ContextPath = "/test"
		})
		g.Expect(e).To(Succeed(), "client with service name should be available")

		urlRewriteCreator := func(ctx context.Context, method string, target *url.URL) (*http.Request, error) {
			if len(target.Port()) != 0 {
				return nil, fmt.Errorf("target URL [%s] should not have port", target.String())
			}
			target.Host = fmt.Sprintf(`%s:%d`, target.Host, webtest.CurrentPort(ctx))
			return http.NewRequestWithContext(ctx, method, target.String(), nil)
		}
		performEchoTest(ctx, t, g, client, httpclient.WithRequestCreator(urlRewriteCreator))
	}
}

func SubTestWithSDNoResponseContent(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithService(SDServiceNameFullInfo)
		g.Expect(e).To(Succeed(), "client with service name should be available")
		performNoResponseBodyTest(ctx, t, g, client)
	}
}

func SubTestWithSDFailure(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		_, e := di.HttpClient.WithService("")
		//checking the error type, ignoring the error message
		g.Expect(e).To(MatchError(httpclient.NewNoEndpointFoundError("error message doesn't matter")))
		//check that the message is formatted
		g.Expect(e).To(MatchError(Not(ContainSubstring("%"))))
	}
}

func SubTestWithBaseURL(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		baseUrl := fmt.Sprintf(`http://localhost:%d%s`, webtest.CurrentPort(ctx), webtest.CurrentContextPath(ctx))
		client, e := di.HttpClient.WithBaseUrl(baseUrl)
		g.Expect(e).To(Succeed(), "client with base URL should be available")
		performEchoTest(ctx, t, g, client)
	}
}

func SubTestWithBaseURLFailure(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		baseUrl := webtest.CurrentContextPath(ctx)
		_, e := di.HttpClient.WithBaseUrl(baseUrl)
		//checking the error type, ignoring the error message
		g.Expect(e).To(MatchError(httpclient.NewNoEndpointFoundError("error message doesn't matter")))
		//check that the message is formatted
		g.Expect(e).To(MatchError(Not(ContainSubstring("%"))))
		g.Expect(e).To(MatchError(ContainSubstring(baseUrl)))
	}
}

func SubTestWithErrorResponse(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithService(SDServiceNameFullInfo)
		g.Expect(e).To(Succeed(), "client with service name should be available")

		sc := 400 + utils.RandomIntN(10)
		reqBody := makeEchoRequestBody()
		req := httpclient.NewRequest(TestErrorPath, http.MethodPut,
			httpclient.WithParam("sc", fmt.Sprintf("%d", sc)),
			httpclient.WithBody(reqBody),
		)

		_, err := client.Execute(ctx, req, httpclient.JsonBody(&EchoResponse{}))
		assertErrorResponse(t, g, err, sc)
	}
}

func SubTestWithUnexpectedStatusCodeInErrorResponse(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithService(SDServiceNameFullInfo)
		g.Expect(e).To(Succeed(), "client with service name should be available")

		expectedErrorForSC := map[int]error{
			300: httpclient.ErrorSubTypeInternalError,
			301: httpclient.ErrorSubTypeInternalError,
			// in our test, the mock server doesn't give a valid redirect location,
			// so we expect error response
			302: httpclient.ErrorSubTypeInternalError,
			303: httpclient.ErrorSubTypeInternalError,
			307: httpclient.ErrorSubTypeInternalError,
			308: httpclient.ErrorSubTypeInternalError,
			// when 304 is returned, the response will not contain content type header,
			// therefore it will result in a media type error
			304: httpclient.ErrorSubTypeMedia,
		}

		for sc, expectedErr := range expectedErrorForSC {
			reqBody := makeEchoRequestBody()
			req := httpclient.NewRequest(TestErrorPath, http.MethodPut,
				httpclient.WithParam("sc", fmt.Sprintf("%d", sc)),
				httpclient.WithBody(reqBody),
			)

			_, err := client.Execute(ctx, req, httpclient.JsonBody(&EchoResponse{}))
			// check the error type is the expected type
			g.Expect(err).To(MatchError(expectedErr))
			// check the error message is formatted
			g.Expect(err).To(MatchError(Not(ContainSubstring("%"))))
		}
	}
}

func SubTestWithNoContentErrorResponse(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithService(SDServiceNameFullInfo)
		g.Expect(e).To(Succeed(), "client with service name should be available")

		sc := 400 + utils.RandomIntN(10)
		req := httpclient.NewRequest(TestNoContentErrorPath, http.MethodPut,
			httpclient.WithParam("sc", fmt.Sprintf("%d", sc)),
		)

		_, err := client.Execute(ctx, req, httpclient.JsonBody(&NoContentResponse{}))
		assertNoContentErrorResponse(t, g, err, sc)
	}
}

func SubTestWithFailedSD(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithService("non-existing")
		g.Expect(e).To(Succeed(), "client with service name should be available")

		req := httpclient.NewRequest(TestPath, http.MethodGet)
		_, err := client.Execute(ctx, req, httpclient.JsonBody(&EchoResponse{}))
		g.Expect(err).To(HaveOccurred(), "execute request with non-existing service should fail")
		g.Expect(errors.Is(err, httpclient.ErrorSubTypeDiscovery)).To(BeTrue(), "error should have correct type")
		g.Expect(err).To(BeAssignableToTypeOf(&httpclient.Error{}), "error should be correct type")
	}
}

func SubTestWithRetry(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithConfig(&httpclient.ClientConfig{
			MaxRetries:   1,                // this should be overridden by RetryCallback
			RetryBackoff: 10 * time.Minute, // this should be overridden by RetryCallback
			Timeout:      5 * time.Second,
			RetryCallback: func(n int, _ interface{}, _ error) (shouldContinue bool, backoff time.Duration) {
				return n < 3, 10 * time.Millisecond
			},
		}).WithService(SDServiceNameFullInfo)
		g.Expect(e).To(Succeed(), "client with service name should be available")

		// do test with success rate 1 per 3 requests
		di.MockedController.Count = 0
		sc := 400 + utils.RandomIntN(10)
		reqBody := makeEchoRequestBody()
		req := httpclient.NewRequest(TestMaybeFailPath, http.MethodPut,
			httpclient.WithParam("sc", fmt.Sprintf("%d", sc)),
			httpclient.WithParam("rate", "3"),
			httpclient.WithBody(reqBody),
		)
		resp, e := client.Execute(ctx, req, httpclient.JsonBody(&NoContentResponse{}))
		g.Expect(e).To(Succeed(), "execute should not fail")
		assertNoContentResponse(t, g, resp, http.StatusOK)

		// do test with success rate 1 per 4 requests
		di.MockedController.Count = 0
		req = httpclient.NewRequest(TestMaybeFailPath, http.MethodPut,
			httpclient.WithParam("sc", fmt.Sprintf("%d", sc)),
			httpclient.WithParam("rate", "4"),
			httpclient.WithBody(reqBody),
		)
		_, e = client.Execute(ctx, req, httpclient.JsonBody(&NoContentResponse{}))
		g.Expect(e).To(HaveOccurred(), "execute should fail")
	}
}

func SubTestWithTimeout(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithConfig(&httpclient.ClientConfig{
			RetryBackoff: time.Minute,
			Timeout:      10 * time.Millisecond,
		}).WithService(SDServiceNameFullInfo)
		g.Expect(e).To(Succeed(), "client with service name should be available")

		// 1st attempt fails, 2nd attempt success, but should timeout before 2nd attempt
		di.MockedController.Count = 0
		sc := 400 + utils.RandomIntN(10)
		reqBody := makeEchoRequestBody()
		req := httpclient.NewRequest(TestMaybeFailPath, http.MethodPut,
			httpclient.WithParam("sc", fmt.Sprintf("%d", sc)),
			httpclient.WithParam("rate", "2"),
			httpclient.WithBody(reqBody),
		)
		_, e = client.Execute(ctx, req, httpclient.JsonBody(&NoContentResponse{}))
		g.Expect(e).To(HaveOccurred(), "execution should return error")
		assertErrorResponse(t, g, e, sc)

		// first attempt times out
		req = httpclient.NewRequest(TestTimeoutPath, http.MethodPost, httpclient.WithBody(reqBody))
		_, e = client.Execute(ctx, req, httpclient.JsonBody(&NoContentResponse{}))
		g.Expect(e).To(HaveOccurred(), "execution should return error")
		g.Expect(e).To(gomegautils.IsError(httpclient.ErrorSubTypeTimeout), "error should be correct type")
	}
}

func SubTestWithURLEncoded(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithService(SDServiceNameFullInfo)
		g.Expect(e).To(Succeed(), "client with service name should be available")

		body := url.Values{}
		random := utils.RandomString(20)
		now := time.Now().Format(time.RFC3339)
		body.Set("data", random)
		req := httpclient.NewRequest(TestPath, http.MethodPost,
			httpclient.WithHeader("X-Data", random),
			httpclient.WithParam("time", now),
			httpclient.WithUrlEncodedBody(body),
		)

		resp, e := client.Execute(ctx, req, httpclient.JsonBody(&EchoResponse{}))
		g.Expect(e).To(Succeed(), "execute request shouldn't fail")

		expected := EchoResponse{
			Headers: map[string]string{
				"X-Data": random,
			},
			Form: map[string]string{
				"time": now,
				"data": random,
			},
		}
		assertResponse(t, g, resp, http.StatusOK, &expected)
	}
}

func SubTestWithAbsoluteUrl(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithNoTargetResolver()
		g.Expect(e).To(Succeed(), "client without target resolver should be available")

		random := utils.RandomString(20)
		now := time.Now().Format(time.RFC3339)
		reqBody := makeEchoRequestBody()
		opts := append([]httpclient.RequestOptions{
			httpclient.WithHeader("X-Data", random),
			httpclient.WithParam("time", now),
			httpclient.WithParam("data", random),
			httpclient.WithBody(reqBody),
		})

		uri, e := url.Parse(fmt.Sprintf(`http://localhost:%d%s`, webtest.CurrentPort(ctx), webtest.CurrentContextPath(ctx)))
		g.Expect(e).ToNot(HaveOccurred())

		uri.Path = path.Join(uri.Path, TestPath)
		req := httpclient.NewRequest(uri.String(), http.MethodPost, opts...)

		resp, e := client.Execute(ctx, req, httpclient.JsonBody(&EchoResponse{}))
		g.Expect(e).To(Succeed(), "execute request shouldn't fail")

		expected := EchoResponse{
			Headers: map[string]string{
				"X-Data": random,
			},
			Form: map[string]string{
				"time": now,
				"data": random,
			},
			ReqBody: reqBody,
		}
		assertResponse(t, g, resp, http.StatusOK, &expected)
	}
}

/*************************
	Request/Response
 *************************/

type Payload struct {
	String string `json:"string"`
	Number int    `json:"number"`
	Bool   bool   `json:"bool"`
}

type EchoRequest struct {
	Payload
	Array  []Payload `json:"array"`
	Object Payload   `json:"object"`
}

type EchoResponse struct {
	Headers map[string]string `json:"headers"`
	Form    map[string]string `json:"form"`
	ReqBody EchoRequest       `json:"body"`
}

type NoContentResponse struct {
	Headers map[string]string `json:"headers"`
	Form    map[string]string `json:"form"`
}

/*************************
	internal
 *************************/

func makeRandomPayload() Payload {
	return Payload{
		String: utils.RandomString(10),
		Number: utils.RandomIntN(1000),
		Bool:   utils.RandomIntN(1000)%2 != 0,
	}
}

func makeEchoRequestBody() EchoRequest {
	return EchoRequest{
		Payload: makeRandomPayload(),
		Array: []Payload{
			makeRandomPayload(), makeRandomPayload(),
		},
		Object: makeRandomPayload(),
	}
}

func performEchoTest(ctx context.Context, t *testing.T, g *gomega.WithT, client httpclient.Client, reqOpts ...httpclient.RequestOptions) {
	random := utils.RandomString(20)
	now := time.Now().Format(time.RFC3339)
	reqBody := makeEchoRequestBody()
	opts := append([]httpclient.RequestOptions{
		httpclient.WithHeader("X-Data", random),
		httpclient.WithParam("time", now),
		httpclient.WithParam("data", random),
		httpclient.WithBody(reqBody),
	}, reqOpts...)
	req := httpclient.NewRequest(TestPath, http.MethodPost, opts...)

	resp, e := client.Execute(ctx, req, httpclient.JsonBody(&EchoResponse{}))
	g.Expect(e).To(Succeed(), "execute request shouldn't fail")

	expected := EchoResponse{
		Headers: map[string]string{
			"X-Data": random,
		},
		Form: map[string]string{
			"time": now,
			"data": random,
		},
		ReqBody: reqBody,
	}
	assertResponse(t, g, resp, http.StatusOK, &expected)
}

func performNoResponseBodyTest(ctx context.Context, t *testing.T, g *gomega.WithT, client httpclient.Client, reqOpts ...httpclient.RequestOptions) {
	random := utils.RandomString(20)
	now := time.Now().Format(time.RFC3339)
	opts := append([]httpclient.RequestOptions{
		httpclient.WithHeader("X-Data", random),
		httpclient.WithParam("time", now),
		httpclient.WithParam("data", random),
	}, reqOpts...)
	req := httpclient.NewRequest(TestNoContentPath, http.MethodPost, opts...)

	resp, e := client.Execute(ctx, req, httpclient.JsonBody(&NoContentResponse{}))
	g.Expect(e).To(Succeed(), "execute request shouldn't fail")
	assertNoContentResponse(t, g, resp, http.StatusNoContent)
}

func assertResponse(_ *testing.T, g *gomega.WithT, resp *httpclient.Response, expectedSC int, expectedBody *EchoResponse) {
	g.Expect(resp).To(Not(BeNil()), "response cannot be nil")
	g.Expect(resp.StatusCode).To(Equal(expectedSC), "response status code should be correct")
	g.Expect(resp.Headers).To(HaveKey("Content-Type"), "response headers should at least have content-type")
	g.Expect(resp.RawBody).To(Not(BeEmpty()), "Response.RawBody shouldn't be empty")
	g.Expect(resp.Body).To(BeAssignableToTypeOf(expectedBody), "Response.Body should be correct type")

	respBody := resp.Body.(*EchoResponse)
	for k, v := range expectedBody.Headers {
		if len(v) != 0 {
			g.Expect(respBody.Headers).To(HaveKeyWithValue(k, v), ".Headers should contains %s=%s", k, v)
		} else {
			g.Expect(respBody.Headers).ToNot(HaveKey(k), ".Headers should not contains %s", k)
		}
	}

	for k, v := range expectedBody.Form {
		if len(v) != 0 {
			g.Expect(respBody.Form).To(HaveKeyWithValue(k, v), ".Form should contains %s=%s", k, v)
		} else {
			g.Expect(respBody.Form).ToNot(HaveKey(k), ".Form should not contains %s", k)
		}
	}

	g.Expect(respBody.ReqBody).To(BeEquivalentTo(expectedBody.ReqBody), ".ReqBody should correct")
}

func assertNoContentResponse(_ *testing.T, g *gomega.WithT, resp *httpclient.Response, expectedSC int) {
	g.Expect(resp).To(Not(BeNil()), "response cannot be nil")
	g.Expect(resp.StatusCode).To(Equal(expectedSC), "response status code should be correct")
	g.Expect(resp.Headers).To(HaveKey("Content-Type"), "response headers should at least have content-type")
}

func assertNoContentErrorResponse(_ *testing.T, g *gomega.WithT, err error, expectedSC int) {
	g.Expect(err).To(HaveOccurred(), "execute request with random values should fail")
	g.Expect(err).To(BeAssignableToTypeOf(&httpclient.Error{}), "error should be correct type")

	resp := err.(*httpclient.Error).Response
	g.Expect(resp).To(Not(BeNil()), "error should contains response")
	g.Expect(resp.StatusCode).To(Equal(expectedSC), "error response should have correct status code")
	g.Expect(resp.Header).To(HaveKey("Content-Type"), "error response headers should at least have content-type")
	g.Expect(resp.Body).To(Not(BeNil()), "error response should have parsed body")
}

func assertErrorResponse(_ *testing.T, g *gomega.WithT, err error, expectedSC int) {
	g.Expect(err).To(HaveOccurred(), "execute request with random values should fail")
	g.Expect(err.Error()).To(HaveSuffix(TestErrorMessageSuffix), "error should have correct value")
	g.Expect(errors.Is(err, httpclient.ErrorTypeResponse)).To(BeTrue(), "error should have correct type")
	g.Expect(err).To(BeAssignableToTypeOf(&httpclient.Error{}), "error should be correct type")

	resp := err.(*httpclient.Error).Response
	g.Expect(resp).To(Not(BeNil()), "error should contains response")
	g.Expect(resp.StatusCode).To(Equal(expectedSC), "error response should have correct status code")
	g.Expect(resp.Header).To(HaveKey("Content-Type"), "error response headers should at least have content-type")
	g.Expect(resp.RawBody).To(Not(BeEmpty()), "error response shouldn't be empty")
	g.Expect(resp.Error()).To(HaveSuffix(TestErrorMessageSuffix), "error response should correct error field")
	g.Expect(resp.Message()).To(HaveSuffix(TestErrorMessageSuffix), "error response should correct message")

	g.Expect(resp.Body).To(Not(BeNil()), "error response should have parsed body")
	g.Expect(resp.Body.Error()).To(HaveSuffix(TestErrorMessageSuffix), "error response should correct error field")
	g.Expect(resp.Body.Message()).To(HaveSuffix(TestErrorMessageSuffix), "error response should correct message")

	//g.Expect(resp.StatusCode).To(Equal(expectedSC), "response status code should be correct")
	//
	//respBody := resp.Body.(*EchoResponse)
	//for k, v := range expectedBody.Headers {
	//	g.Expect(respBody.Headers).To(HaveKeyhttpclient.WithValue(k, v), ".Headers should contains %s=%s", k, v)
	//}
	//
	//for k, v := range expectedBody.Form {
	//	g.Expect(respBody.Form).To(HaveKeyhttpclient.WithValue(k, v), ".Headers should contains %s=%s", k, v)
	//}
	//
	//g.Expect(respBody.ReqBody).To(BeEquivalentTo(expectedBody.ReqBody), ".ReqBody should correct")
}

//func assertResult(_ *testing.T, g *gomega.WithT, i interface{}, expectedUser string) {
//	g.Expect(i).To(Not(BeNil()), "functions that calling remote service should have proper response")
//	g.Expect(i).To(BeAssignableToTypeOf(map[string]interface{}{}), "service should return a map")
//	m := i.(map[string]interface{})
//	g.Expect(m).To(HaveKeyhttpclient.WithValue("username", expectedUser), "body should contains correct username")
//}
