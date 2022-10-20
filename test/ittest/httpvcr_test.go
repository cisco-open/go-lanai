package ittest

import (
	"context"
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
	"net/http"
	"testing"
)

const (
	RecordName     = `HttpVCRTestRecords`
	PathGet        = `/knock`
	PathPost       = `/knock`
	JsonKeyMessage = "message"
	RespMessage    = "this should be recorded"
)

/*************************
	Setup
 *************************/

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

func (c TestController) Get(_ context.Context, _ *http.Request) (interface{}, error) {
	return map[string]interface{}{
		JsonKeyMessage: RespMessage,
	}, nil
}

func (c TestController) Post(_ context.Context, _ *http.Request) (interface{}, error) {
	return map[string]interface{}{
		JsonKeyMessage: RespMessage,
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
		WithHttpPlayback(t, HttpRecordName(RecordName), EnableHttpRecordMode()),
		apptest.WithModules(),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			web.FxControllerProviders(NewTestController),
		),
		test.GomegaSubTest(SubTestHttpVCR(&di), "TestHttpVCRRecording"),
	)
}

func TestHttpVCRPlayback(t *testing.T) {
	var di vcrDI
	t.Name()
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), HttpRecordIgnoreHost()),
		apptest.WithModules(),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestHttpVCR(&di), "TestHttpVCRReplay"),
		test.GomegaSubTest(SubTestHttpVCRIncorrectRequestOrder(&di), "TestHttpVCRIncorrectRequestOrder"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestHttpVCR(di *vcrDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")
		var req *http.Request
		var resp *http.Response
		var e error

		req = newGetRequest(ctx, t, g)
		resp, e = di.Recorder.GetDefaultClient().Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")

		req = newPostRequest(ctx, t, g)
		resp, e = di.Recorder.GetDefaultClient().Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")
	}
}

func SubTestHttpVCRIncorrectRequestOrder(di *vcrDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")
		var req *http.Request
		var e error

		req = newPostRequest(ctx, t, g)
		_, e = di.Recorder.GetDefaultClient().Do(req)
		g.Expect(e).To(HaveOccurred(), "sending request in wrong order should fail")

		req = newGetRequest(ctx, t, g)
		_, e = di.Recorder.GetDefaultClient().Do(req)
		g.Expect(e).To(HaveOccurred(), "sending request in wrong order should fail")
	}
}

/*************************
	internal
 *************************/

func newGetRequest(ctx context.Context, _ *testing.T, g *gomega.WithT) *http.Request {
	port := webtest.CurrentPort(ctx)
	if port < 0 {
		port = 8080
	}
	url := fmt.Sprintf("http://localhost:%d%s%s", port, webtest.CurrentContextPath(ctx), PathGet)
	req, e := http.NewRequest(http.MethodGet, url, nil)
	g.Expect(e).To(Succeed(), "creating request should succeed")
	return req
}

func newPostRequest(ctx context.Context, _ *testing.T, g *gomega.WithT) *http.Request {
	port := webtest.CurrentPort(ctx)
	if port < 0 {
		port = 8080
	}
	url := fmt.Sprintf("http://localhost:%d%s%s", port, webtest.CurrentContextPath(ctx), PathPost)
	req, e := http.NewRequest(http.MethodPost, url, nil)
	g.Expect(e).To(Succeed(), "creating request should succeed")
	return req
}