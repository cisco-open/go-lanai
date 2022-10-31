package httpclient

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sdtest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"errors"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
	"time"
)

/*************************
	Setup
 *************************/

const (
	SDServiceName = `mockedserver`
	TestPath      = "/echo"
	TestErrorPath = "/fail"
)

// UpdateMockedSD update SD record to use the random server port
func UpdateMockedSD(di *TestDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		port := webtest.CurrentPort(ctx)
		if port <= 0 {
			return ctx, nil
		}
		di.Client.UpdateMockedService(SDServiceName, sdtest.NthInstance(0), func(inst *discovery.Instance) {
			inst.Port = port
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
	HttpClient Client
}

func TestExampleMockedServerTestWithSecurity(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		sdtest.WithMockedSD(sdtest.DefinitionWithPrefix("mocks.sd")),
		apptest.WithModules(Module),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			web.FxControllerProviders(testdata.NewMockedController),
		),
		test.SubTestSetup(UpdateMockedSD(&di)),
		test.GomegaSubTest(SubTestWithSD(&di), "TestWithSD"),
		test.GomegaSubTest(SubTestWithBaseURL(&di), "TestWithBaseURL"),
		test.GomegaSubTest(SubTestWithErrorResponse(&di), "TestWithErrorResponse"),
		test.GomegaSubTest(SubTestWithFailedSD(&di), "TestWithFailedSD"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestWithSD(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithService(SDServiceName)
		g.Expect(e).To(Succeed(), "client with service name should be available")
		performEchoTest(ctx, t, g, client)
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

func SubTestWithErrorResponse(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithService("mockedserver")
		g.Expect(e).To(Succeed(), "client with service name should be available")

		sc := 400 + utils.RandomIntN(10)
		reqBody := makeEchoRequest()
		req := NewRequest(TestErrorPath, http.MethodPut,
			WithParam("sc", fmt.Sprintf("%d", sc)),
			WithBody(reqBody),
		)

		_, err := client.Execute(ctx, req, JsonBody(&EchoResponse{}))
		assertErrorResponse(t, g, err, sc)
	}
}

func SubTestWithFailedSD(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithService("non-existing")
		g.Expect(e).To(Succeed(), "client with service name should be available")

		req := NewRequest(TestPath, http.MethodGet)
		_, err := client.Execute(ctx, req, JsonBody(&EchoResponse{}))
		g.Expect(err).To(HaveOccurred(), "execute request with non-existing service should fail")
		g.Expect(errors.Is(err, ErrorSubTypeDiscovery)).To(BeTrue(), "error should have correct type")
		g.Expect(err).To(BeAssignableToTypeOf(&Error{}), "error should be correct type")
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

func makeEchoRequest() EchoRequest {
	return EchoRequest{
		Payload: makeRandomPayload(),
		Array: []Payload{
			makeRandomPayload(), makeRandomPayload(),
		},
		Object: makeRandomPayload(),
	}
}

func performEchoTest(ctx context.Context, t *testing.T, g *gomega.WithT, client Client) {
	random := utils.RandomString(20)
	now := time.Now().Format(time.RFC3339)
	reqBody := makeEchoRequest()
	req := NewRequest(TestPath, http.MethodPost,
		WithHeader("X-Data", random),
		WithParam("time", now),
		WithParam("data", random),
		WithBody(reqBody),
	)

	resp, e := client.Execute(ctx, req, JsonBody(&EchoResponse{}))
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

func assertResponse(_ *testing.T, g *gomega.WithT, resp *Response, expectedSC int, expectedBody *EchoResponse) {
	g.Expect(resp).To(Not(BeNil()), "response cannot be nil")
	g.Expect(resp.StatusCode).To(Equal(expectedSC), "response status code should be correct")
	g.Expect(resp.Headers).To(HaveKey("Content-Type"), "response headers should at least have content-type")
	g.Expect(resp.RawBody).To(Not(BeEmpty()), "Response.RawBody shouldn't be empty")
	g.Expect(resp.Body).To(BeAssignableToTypeOf(expectedBody), "Response.Body should be correct type")

	respBody := resp.Body.(*EchoResponse)
	for k, v := range expectedBody.Headers {
		g.Expect(respBody.Headers).To(HaveKeyWithValue(k, v), ".Headers should contains %s=%s", k, v)
	}

	for k, v := range expectedBody.Form {
		g.Expect(respBody.Form).To(HaveKeyWithValue(k, v), ".Headers should contains %s=%s", k, v)
	}

	g.Expect(respBody.ReqBody).To(BeEquivalentTo(expectedBody.ReqBody), ".ReqBody should correct")
}

func assertErrorResponse(_ *testing.T, g *gomega.WithT, err error, expectedSC int) {
	g.Expect(err).To(HaveOccurred(), "execute request with random values should fail")
	g.Expect(err.Error()).To(HaveSuffix(testdata.ErrorMessage), "error should have correct value")
	g.Expect(errors.Is(err, ErrorTypeResponse)).To(BeTrue(), "error should have correct type")
	g.Expect(err).To(BeAssignableToTypeOf(&Error{}), "error should be correct type")

	resp := err.(*Error).Response
	g.Expect(resp).To(Not(BeNil()), "error should contains response")
	g.Expect(resp.StatusCode).To(Equal(expectedSC), "error response should have correct status code")
	g.Expect(resp.Header).To(HaveKey("Content-Type"), "error response headers should at least have content-type")
	g.Expect(resp.RawBody).To(Not(BeEmpty()), "error response shouldn't be empty")
	g.Expect(resp.Error()).To(Equal(testdata.ErrorMessage), "error response should correct error field")
	g.Expect(resp.Message()).To(Equal(testdata.ErrorMessage), "error response should correct message")

	g.Expect(resp.Body).To(Not(BeNil()), "error response should have parsed body")
	g.Expect(resp.Body.Error()).To(Equal(testdata.ErrorMessage), "error response should correct error field")
	g.Expect(resp.Body.Message()).To(Equal(testdata.ErrorMessage), "error response should correct message")


	//g.Expect(resp.StatusCode).To(Equal(expectedSC), "response status code should be correct")
	//
	//respBody := resp.Body.(*EchoResponse)
	//for k, v := range expectedBody.Headers {
	//	g.Expect(respBody.Headers).To(HaveKeyWithValue(k, v), ".Headers should contains %s=%s", k, v)
	//}
	//
	//for k, v := range expectedBody.Form {
	//	g.Expect(respBody.Form).To(HaveKeyWithValue(k, v), ".Headers should contains %s=%s", k, v)
	//}
	//
	//g.Expect(respBody.ReqBody).To(BeEquivalentTo(expectedBody.ReqBody), ".ReqBody should correct")
}

//func assertResult(_ *testing.T, g *gomega.WithT, i interface{}, expectedUser string) {
//	g.Expect(i).To(Not(BeNil()), "functions that calling remote service should have proper response")
//	g.Expect(i).To(BeAssignableToTypeOf(map[string]interface{}{}), "service should return a map")
//	m := i.(map[string]interface{})
//	g.Expect(m).To(HaveKeyWithValue("username", expectedUser), "body should contains correct username")
//}
