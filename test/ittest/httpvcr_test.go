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

type vcrTestDI struct {
	fx.In
	Recorder *recorder.Recorder
}

func TestHttpVCRRecording(t *testing.T) {
	var di vcrTestDI
	t.Name()
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		WithHttpPlayback(t, HttpRecordName(RecordName),
			HttpRecordingMode(), HttpRecorderHooks(NewRecorderHookWithOrder(LocalhostRewriteHook(), recorder.BeforeSaveHook, 0))),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			web.FxControllerProviders(NewTestController),
		),
		test.GomegaSubTest(SubTestHttpVCRMode(true), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestNormalGet(&di), "TestHttpVCRRecordGet"),
		test.GomegaSubTest(SubTestNormalPostJson(&di), "TestHttpVCRRecordPostJson"),
		test.GomegaSubTest(SubTestNormalPostForm(&di), "TestHttpVCRRecordPostForm"),
	)
}

func TestHttpVCRPlaybackExact(t *testing.T) {
	var di vcrTestDI
	t.Name()
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), DisableHttpRecordingMode(), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestNormalGet(&di), "TestNormalGet"),
		test.GomegaSubTest(SubTestNormalPostJson(&di), "TestNormalPostJson"),
		test.GomegaSubTest(SubTestNormalPostForm(&di), "TestNormalPostForm"),
	)
}

func TestHttpVCRPlaybackEquivalent(t *testing.T) {
	var di vcrTestDI
	t.Name()
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), DisableHttpRecordingMode(), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestEquivalentGet(&di), "TestEquivalentGet"),
		test.GomegaSubTest(SubTestEquivalentPostJson(&di), "TestEquivalentPostJson"),
		test.GomegaSubTest(SubTestEquivalentPostForm(&di), "TestEquivalentPostForm"),
	)
}

func TestHttpVCRPlaybackIncorrectOrder(t *testing.T) {
	var di vcrTestDI
	t.Name()
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), DisableHttpRecordingMode(), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestDifferentRequestOrder(&di, false), "TestDifferentRequestOrder"),
	)
}

func TestHttpVCRPlaybackWithOrderDisabled(t *testing.T) {
	var di vcrTestDI
	t.Name()
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), DisableHttpRecordingMode(), HttpRecordIgnoreHost(), DisableHttpRecordOrdering()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestDifferentRequestOrder(&di, true), "TestDifferentRequestOrder"),
	)
}

func TestHttpVCRPlaybackIncorrectQuery(t *testing.T) {
	var di vcrTestDI
	t.Name()
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), DisableHttpRecordingMode(), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestIncorrectRequestQuery(&di), "TestIncorrectRequestQuery"),
	)
}

func TestHttpVCRPlaybackIncorrectBody(t *testing.T) {
	var di vcrTestDI
	t.Name()
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), DisableHttpRecordingMode(), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestIncorrectRequestJsonBody(&di), "TestIncorrectRequestJsonBody"),
		test.GomegaSubTest(SubTestIncorrectRequestFormBody(&di), "TestIncorrectRequestFormBody"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestHttpVCRMode(expectRecording bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(IsRecording(ctx)).To(Equal(expectRecording), "mode should be correct")
	}
}

func SubTestNormalGet(di *vcrTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")
		req := newGetRequest(ctx, t, g)
		resp, e := Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")
	}
}

func SubTestNormalPostJson(di *vcrTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")
		req := newPostJsonRequest(ctx, t, g)
		resp, e := Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")

	}
}

func SubTestNormalPostForm(di *vcrTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")

		req := newPostFormRequest(ctx, t, g)
		resp, e := Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")
	}
}

func SubTestEquivalentGet(di *vcrTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")
		req := newGetRequest(ctx, t, g)
		resp, e := Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")
	}
}

func SubTestEquivalentPostJson(di *vcrTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")

		AdditionalMatcherOptions(ctx, FuzzyJsonPaths("$.time"))
		req := newPostJsonRequest(ctx, t, g, withBody(CorrectRequestJsonBodyAlt))
		resp, e := Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")

	}
}

func SubTestEquivalentPostForm(di *vcrTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")

		AdditionalMatcherOptions(ctx, FuzzyForm("time"))
		req := newPostFormRequest(ctx, t, g, withBody(CorrectRequestFormBodyAlt))
		resp, e := Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")
	}
}

func SubTestDifferentRequestOrder(di *vcrTestDI, expectSuccess bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")
		var req *http.Request
		var resp *http.Response
		var e error

		req = newPostJsonRequest(ctx, t, g)
		resp, e = Client(ctx).Do(req)
		if expectSuccess {
			g.Expect(e).To(Succeed(), "sending request in different order should succeed")
			g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")
		} else {
			g.Expect(e).To(HaveOccurred(), "sending request in different order should fail")
		}

		req = newGetRequest(ctx, t, g)
		resp, e = Client(ctx).Do(req)
		if expectSuccess {
			g.Expect(e).To(Succeed(), "sending request in different order should succeed")
			g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")
		} else {
			g.Expect(e).To(HaveOccurred(), "sending request in different order should fail")
		}
	}
}

func SubTestIncorrectRequestQuery(di *vcrTestDI) test.GomegaSubTestFunc {
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

func SubTestIncorrectRequestJsonBody(di *vcrTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")
		req := newPostJsonRequest(ctx, t, g, withBody(IncorrectRequestJsonBody))
		_, e := Client(ctx).Do(req)
		g.Expect(e).To(HaveOccurred(), "sending request with wrong body should fail")
	}
}

func SubTestIncorrectRequestFormBody(di *vcrTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")

		req := newPostFormRequest(ctx, t, g, withBody(IncorrectRequestFormBody))
		_, e := Client(ctx).Do(req)
		g.Expect(e).To(HaveOccurred(), "sending request with wrong body should fail")
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
	req, e := http.NewRequest(http.MethodGet, url, nil)
	g.Expect(e).To(Succeed(), "creating request should succeed")

	prepareRequest(req, "application/x-www-form-urlencoded; charset=utf-8", opts)
	return req
}

func newPostJsonRequest(ctx context.Context, _ *testing.T, g *gomega.WithT, opts ...webtest.RequestOptions) *http.Request {
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

func newPostFormRequest(ctx context.Context, _ *testing.T, g *gomega.WithT, opts ...webtest.RequestOptions) *http.Request {
	port := webtest.CurrentPort(ctx)
	if port < 0 {
		port = 8080
	}
	url := fmt.Sprintf("http://localhost:%d%s%s", port, webtest.CurrentContextPath(ctx), PathPost)
	req, e := http.NewRequest(http.MethodPost, url, strings.NewReader(CorrectRequestFormBody))
	g.Expect(e).To(Succeed(), "creating request should succeed")

	prepareRequest(req, "application/x-www-form-urlencoded; charset=utf-8", opts)
	return req
}

func prepareRequest(req *http.Request, contentType string, opts []webtest.RequestOptions) {
	// set headers
	for k  := range FuzzyRequestHeaders {
		req.Header.Set(k, utils.RandomString(20))
	}
	req.Header.Set("Content-Type", contentType)

	// set sensitive queries
	q := req.URL.Query()
	for k := range FuzzyRequestQueries {
		q.Set(k, utils.RandomString(10))
	}
	q.Set(RequiredQuery, "value should match")
	req.URL.RawQuery = q.Encode()

	for _, fn := range opts {
		fn(req)
	}
}
