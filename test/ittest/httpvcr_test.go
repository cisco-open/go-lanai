package ittest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

/*************************
	Setup
 *************************/

const (
	RecordName    = `HttpVCRTestRecords`
	PathGet       = `/knock`
	PathPost      = `/knock`
	RequiredQuery = "important"
)

const (
	CorrectRequestJsonBody    = `{"string":"correct","number":1,"bool":true,"time":"2022-10-11T00:00:00Z"}`
	CorrectRequestFormBody    = `string=correct&number=1&bool=true&time=2022-10-11T00%3A00%3A00Z`
	CorrectRequestJsonBodyAlt = `{"number":1,"bool":true,"time":"1982-10-11T00:00:00Z","string":"correct"}`
	CorrectRequestFormBodyAlt = `number=1&bool=true&time=1982-10-11T00%3A00%3A00Z&string=correct`
	IncorrectRequestJsonBody  = `{"string":"another","number":1,"bool":true,"time":"1982-10-11T00:00:00Z"}`
	IncorrectRequestFormBody  = `string=incorrect&number=1&bool=true&time=1982-10-11T00%3A00%3A00Z`
)

type TestObject struct {
	String string    `json:"string" form:"string"`
	Number float64   `json:"number" form:"number"`
	Bool   bool      `json:"bool" form:"bool"`
	Time   time.Time `json:"time" form:"time"`
}

type TestRequest struct {
	TestObject
}

type TestResponse struct {
	TestObject
	Slice  []TestObject `json:"slice" form:"slice"`
	Object TestObject   `json:"object" form:"object"`
}

type TestController struct{}

func NewTestController() web.Controller {
	return TestController{}
}

func (c TestController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.New("get").Get(PathGet).EndpointFunc(c.Get).Build(),
		rest.New("post").Post(PathPost).EndpointFunc(c.Post).Build(),
	}
}

func (c TestController) Get(_ context.Context, r *TestRequest) (*TestResponse, error) {
	return &TestResponse{
		TestObject: r.TestObject,
		Slice:      []TestObject{r.TestObject},
		Object:     r.TestObject,
	}, nil
}

func (c TestController) Post(_ context.Context, r *TestRequest) (*TestResponse, error) {
	return &TestResponse{
		TestObject: r.TestObject,
		Slice:      []TestObject{r.TestObject},
		Object:     r.TestObject,
	}, nil
}

/*************************
	Tests
 *************************/

type vcrDI struct {
	fx.In
	Recorder *recorder.Recorder
}

func TestHttpVCRRecording(t *testing.T) {
	var di vcrDI
	t.Name()
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		WithHttpPlayback(t, HttpRecordName(RecordName), EnableHttpRecordMode(), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			web.FxControllerProviders(NewTestController),
		),
		test.GomegaSubTest(SubTestNormalInteraction(&di), "TestHttpVCRRecording"),
		test.GomegaSubTest(SubTestHttpVCRMode(true), "TestHttpVCRMode"),
	)
}

func TestHttpVCRPlayback(t *testing.T) {
	var di vcrDI
	t.Name()
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestNormalInteraction(&di), "TestNormalInteraction"),
	)
}

func TestHttpVCRPlaybackIncorrectQuery(t *testing.T) {
	var di vcrDI
	t.Name()
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestIncorrectRequestQuery(&di), "TestIncorrectRequestQuery"),
	)
}

func TestHttpVCRPlaybackIncorrectOrder(t *testing.T) {
	var di vcrDI
	t.Name()
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestIncorrectRequestOrder(&di), "TestIncorrectRequestOrder"),
	)
}

//func TestHttpVCRPlaybackIncorrectBody(t *testing.T) {
//	var di vcrDI
//	t.Name()
//	test.RunTest(context.Background(), t,
//		apptest.Bootstrap(),
//		WithHttpPlayback(t, HttpRecordName(RecordName), HttpRecordIgnoreHost()),
//		apptest.WithDI(&di),
//		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
//		test.GomegaSubTest(SubTestIncorrectRequestBody(&di), "TestIncorrectRequestBody"),
//	)
//}

/*************************
	Sub Tests
 *************************/

func SubTestHttpVCRMode(expectRecording bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(IsRecording(ctx)).To(Equal(expectRecording), "mode should be correct")
	}
}

func SubTestNormalInteraction(di *vcrDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")
		var req *http.Request
		var resp *http.Response
		var e error

		req = newGetRequest(ctx, t, g)
		resp, e = Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")

		req = newPostRequest(ctx, t, g)
		resp, e = Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")
	}
}

func SubTestIncorrectRequestQuery(di *vcrDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")
		var req *http.Request
		var e error

		req = newGetRequest(ctx, t, g, func(req *http.Request) {
			q := req.URL.Query()
			q.Del(RequiredQuery)
			req.URL.RawQuery = q.Encode()
		})
		_, e = Client(ctx).Do(req)
		g.Expect(e).To(HaveOccurred(), "sending request with wrong form body should fail")
	}
}

func SubTestIncorrectRequestOrder(di *vcrDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")
		var req *http.Request
		var e error

		req = newPostRequest(ctx, t, g)
		_, e = Client(ctx).Do(req)
		g.Expect(e).To(HaveOccurred(), "sending request in wrong order should fail")

		req = newGetRequest(ctx, t, g)
		_, e = Client(ctx).Do(req)
		g.Expect(e).To(HaveOccurred(), "sending request in wrong order should fail")
	}
}

func SubTestIncorrectRequestBody(di *vcrDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")
		var req *http.Request
		var e error

		req = newGetRequest(ctx, t, g, withBody(IncorrectRequestFormBody))
		_, e = Client(ctx).Do(req)
		g.Expect(e).To(HaveOccurred(), "sending request with wrong form body should fail")

		req = newPostRequest(ctx, t, g, withBody(IncorrectRequestJsonBody))
		_, e = Client(ctx).Do(req)
		g.Expect(e).To(HaveOccurred(), "sending request with wrong json body should fail")
	}
}

/*************************
	internal
 *************************/

func withBody(body string) webtest.RequestOptions {
	return func(req *http.Request) {
		req.Body = io.NopCloser(strings.NewReader(body))
	}
}

func newGetRequest(ctx context.Context, _ *testing.T, g *gomega.WithT, opts ...webtest.RequestOptions) *http.Request {
	port := webtest.CurrentPort(ctx)
	if port < 0 {
		port = 8080
	}

	url := fmt.Sprintf("http://localhost:%d%s%s", port, webtest.CurrentContextPath(ctx), PathGet)
	req, e := http.NewRequest(http.MethodGet, url, strings.NewReader(CorrectRequestFormBody))
	g.Expect(e).To(Succeed(), "creating request should succeed")

	prepareRequest(req, "application/x-www-form-urlencoded; charset=utf-8", opts)
	return req
}

func newPostRequest(ctx context.Context, _ *testing.T, g *gomega.WithT, opts ...webtest.RequestOptions) *http.Request {
	port := webtest.CurrentPort(ctx)
	if port < 0 {
		port = 8080
	}
	url := fmt.Sprintf("http://localhost:%d%s%s", port, webtest.CurrentContextPath(ctx), PathPost)
	req, e := http.NewRequest(http.MethodPost, url, strings.NewReader(CorrectRequestJsonBody))
	g.Expect(e).To(Succeed(), "creating request should succeed")

	prepareRequest(req, "application/json; charset=utf-8", opts)
	return req
}

func prepareRequest(req *http.Request, contentType string, opts []webtest.RequestOptions) {
	// set headers
	for _, k := range SensitiveHeaders {
		req.Header.Set(k, utils.RandomString(20))
	}
	req.Header.Set("Content-Type", contentType)

	// set sensitive queries
	q := req.URL.Query()
	for _, k := range SensitiveQueries {
		q.Set(k, utils.RandomString(10))
	}
	q.Set(RequiredQuery, "value should match")
	req.URL.RawQuery = q.Encode()

	for _, fn := range opts {
		fn(req)
	}
}
