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

package ittest

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/rest"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

/*************************
	Setup
 *************************/

const (
	RecordName    = `HttpVCRTestRecords`
	RecordAltName = `HttpVCRTestMoreRecords`
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

func TestHttpVCRRecording(t *testing.T) {
	var di RecorderDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		WithHttpPlayback(t, HttpRecordName(RecordName),
			HttpRecordingMode(),
			HttpRecorderHooks(LocalhostRewriteHook()),
		),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			web.FxControllerProviders(NewTestController),
		),
		test.GomegaSubTest(SubTestVcrDI(&di), "TestDI"),
		test.GomegaSubTest(SubTestHttpVCRMode(true), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestNormalGet(), "TestHttpVCRRecordGet"),
		test.GomegaSubTest(SubTestNormalPostJson(), "TestHttpVCRRecordPostJson"),
		test.GomegaSubTest(SubTestNormalPostForm(), "TestHttpVCRRecordPostForm"),
	)
}

func TestHttpVCRPlaybackExact(t *testing.T) {
	var di RecorderDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), DisableHttpRecordingMode(), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestVcrDI(&di), "TestDI"),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestNormalGet(), "TestNormalGet"),
		test.GomegaSubTest(SubTestNormalPostJson(), "TestNormalPostJson"),
		test.GomegaSubTest(SubTestNormalPostForm(), "TestNormalPostForm"),
	)
}

func TestHttpVCRPlaybackEquivalent(t *testing.T) {
	var di RecorderDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), DisableHttpRecordingMode(), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestVcrDI(&di), "TestDI"),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestEquivalentGet(), "TestEquivalentGet"),
		test.GomegaSubTest(SubTestEquivalentPostJson(), "TestEquivalentPostJson"),
		test.GomegaSubTest(SubTestEquivalentPostForm(), "TestEquivalentPostForm"),
	)
}

func TestHttpVCRPlaybackIncorrectOrder(t *testing.T) {
	var di RecorderDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), DisableHttpRecordingMode(), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestVcrDI(&di), "TestDI"),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestDifferentRequestOrder(false), "TestDifferentRequestOrder"),
	)
}

func TestHttpVCRPlaybackWithOrderDisabled(t *testing.T) {
	var di RecorderDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), DisableHttpRecordingMode(), HttpRecordIgnoreHost(), DisableHttpRecordOrdering()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestVcrDI(&di), "TestDI"),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestDifferentRequestOrder(true), "TestDifferentRequestOrder"),
	)
}

func TestHttpVCRPlaybackIncorrectQuery(t *testing.T) {
	var di RecorderDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), DisableHttpRecordingMode(), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestVcrDI(&di), "TestDI"),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestIncorrectRequestQuery(), "TestIncorrectRequestQuery"),
	)
}

func TestHttpVCRPlaybackIncorrectBody(t *testing.T) {
	var di RecorderDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t, HttpRecordName(RecordName), DisableHttpRecordingMode(), HttpRecordIgnoreHost()),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestVcrDI(&di), "TestDI"),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestIncorrectRequestJsonBody(), "TestIncorrectRequestJsonBody"),
		test.GomegaSubTest(SubTestIncorrectRequestFormBody(), "TestIncorrectRequestFormBody"),
	)
}

// TestHttpVCRRecordingAltUsage test alternative usage by using NewHttpRecorder directly
func TestHttpVCRRecordingAltUsage(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		test.Setup(SetupVCR(
			HttpRecordingMode(),
			HttpRecordName(RecordAltName),
			HttpRecorderHooks(LocalhostRewriteHook()),
			SanitizeHttpRecord(),
			FixedHttpRecordDuration(DefaultHTTPDuration),
			FixedHttpRecordDuration(0),
			FixedHttpRecordDuration(DefaultHTTPDuration),
			HttpTransport(http.DefaultTransport),
		)),
		test.Teardown(TeardownVCR()),
		apptest.WithFxOptions(
			web.FxControllerProviders(NewTestController),
		),
		test.GomegaSubTest(SubTestVcrContext(), "TestVcrContext"),
		test.GomegaSubTest(SubTestHttpVCRMode(true), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestNormalGet(), "TestHttpVCRRecordGet"),
		test.GomegaSubTest(SubTestNormalPostJson(), "TestHttpVCRRecordPostJson"),
		test.GomegaSubTest(SubTestNormalPostForm(), "TestHttpVCRRecordPostForm"),
	)
}

func TestHttpVCRPlaybackAltUsage(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.Setup(SetupVCR(
			HttpRecordName(RecordAltName),
			HttpRecordIgnoreHost(),
			ApplyHttpLatency(),
		)),
		test.Teardown(TeardownVCR()),
		test.GomegaSubTest(SubTestVcrContext(), "TestVcrContext"),
		test.GomegaSubTest(SubTestHttpVCRMode(false), "TestHttpVCRMode"),
		test.GomegaSubTest(SubTestNormalGet(), "TestNormalGet"),
		test.GomegaSubTest(SubTestNormalPostJson(), "TestNormalPostJson"),
		test.GomegaSubTest(SubTestNormalPostForm(), "TestNormalPostForm"),
	)
}

func TestHttpVCRRecordsConversion(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestConvertV1ToV2(), "TestConvertV1ToV2"),
	)
}

/*************************
	Sub Tests
 *************************/

func SetupVCR(opts ...HTTPVCROptions) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		opts = append([]HTTPVCROptions{HttpRecordName(t.Name())}, opts...)
		return ContextWithNewHttpRecorder(ctx, opts...)
	}
}

func TeardownVCR() test.TeardownFunc {
	return func(ctx context.Context, t *testing.T) error {
		return StopRecorder(ctx)
	}
}

func SubTestVcrDI(di *RecorderDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).To(Not(BeNil()), "Recorder should be injected")
		g.Expect(di.RecorderOption).To(Not(BeZero()), "RecorderOption should be injected")
		g.Expect(di.RecorderMatcher).To(Not(BeZero()), "RecorderMatcher should be injected")
		g.Expect(di.HTTPVCROption).To(Not(BeZero()), "HTTPVCROption should be injected")
		g.Expect(Recorder(ctx)).To(Not(BeNil()), "Recorder from context should be available")
		if IsRecording(ctx) {
			g.Expect(di.HTTPVCROption.Hooks).To(HaveLen(4), "HTTPVCROption.Hooks should have correct length")
		} else {
			g.Expect(di.HTTPVCROption.Hooks).To(HaveLen(3), "HTTPVCROption.Hooks should have correct length")
		}
	}
}

func SubTestVcrContext() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		rec := Recorder(ctx)
		g.Expect(rec).To(Not(BeNil()), "Recorder from context should be available")
		g.Expect(rec.RawOptions).To(Not(BeZero()), "RawOptions should be available")
		g.Expect(rec.InitMatcher).To(Not(BeZero()), "InitMatcher should be available")
		g.Expect(rec.Options).To(Not(BeZero()), "Options should be available")
		if IsRecording(ctx) {
			g.Expect(rec.Options.Hooks).To(HaveLen(4), "Options.Hooks should have correct length")
		} else {
			g.Expect(rec.Options.Hooks).To(HaveLen(1), "Options.Hooks should have correct length")
		}
	}
}

func SubTestHttpVCRMode(expectRecording bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(IsRecording(ctx)).To(Equal(expectRecording), "mode should be correct")
	}
}

func SubTestNormalGet() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(Recorder(ctx)).To(Not(BeNil()), "Recorder from context should be available")
		req := newGetRequest(ctx, t, g)
		resp, e := Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")
	}
}

func SubTestNormalPostJson() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(Recorder(ctx)).To(Not(BeNil()), "Recorder from context should be available")
		req := newPostJsonRequest(ctx, t, g)
		resp, e := Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")

	}
}

func SubTestNormalPostForm() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(Recorder(ctx)).To(Not(BeNil()), "Recorder from context should be available")

		req := newPostFormRequest(ctx, t, g)
		resp, e := Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")
	}
}

func SubTestEquivalentGet() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(Recorder(ctx)).To(Not(BeNil()), "Recorder from context should be available")
		req := newGetRequest(ctx, t, g)
		resp, e := Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")
	}
}

func SubTestEquivalentPostJson() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(Recorder(ctx)).To(Not(BeNil()), "Recorder from context should be available")

		AdditionalMatcherOptions(ctx, FuzzyJsonPaths("$.time"))
		req := newPostJsonRequest(ctx, t, g, withBody(CorrectRequestJsonBodyAlt))
		resp, e := Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")

	}
}

func SubTestEquivalentPostForm() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(Recorder(ctx)).To(Not(BeNil()), "Recorder from context should be available")

		AdditionalMatcherOptions(ctx, FuzzyForm("time"))
		req := newPostFormRequest(ctx, t, g, withBody(CorrectRequestFormBodyAlt))
		resp, e := Client(ctx).Do(req)
		g.Expect(e).To(Succeed(), "sending request should succeed")
		g.Expect(resp.StatusCode).To(BeEquivalentTo(http.StatusOK), "server should return 200")
	}
}

func SubTestDifferentRequestOrder(expectSuccess bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(Recorder(ctx)).To(Not(BeNil()), "Recorder from context should be available")
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

func SubTestIncorrectRequestQuery() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(Recorder(ctx)).To(Not(BeNil()), "Recorder from context should be available")
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

func SubTestIncorrectRequestJsonBody() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(Recorder(ctx)).To(Not(BeNil()), "Recorder from context should be available")
		req := newPostJsonRequest(ctx, t, g, withBody(IncorrectRequestJsonBody))
		_, e := Client(ctx).Do(req)
		g.Expect(e).To(HaveOccurred(), "sending request with wrong body should fail")
	}
}

func SubTestIncorrectRequestFormBody() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(Recorder(ctx)).To(Not(BeNil()), "Recorder from context should be available")

		req := newPostFormRequest(ctx, t, g, withBody(IncorrectRequestFormBody))
		_, e := Client(ctx).Do(req)
		g.Expect(e).To(HaveOccurred(), "sending request with wrong body should fail")
	}
}

func SubTestConvertV1ToV2() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const destName = `testdata/.tmp/v1-to-v2-result.httpvcr`
		e := os.MkdirAll("testdata/.tmp/", 0755)
		g.Expect(e).To(Succeed(), "prepare .tmp folder should not fail")

		destPath := fmt.Sprintf(`%s.yaml`, destName)
		e = ConvertCassetteFileV1toV2("testdata/V1Records.httpvcr.yaml", destPath)
		g.Expect(e).To(Succeed(), "converting v1 to v2 should not fail")

		rec, e := recorder.NewWithOptions(&recorder.Options{
			CassetteName:       destName,
			Mode:               recorder.ModeReplayOnly,
			SkipRequestLatency: false,
		})
		g.Expect(e).To(Succeed(), "create recorder with converted cassette file should not fail")
		g.Expect(rec).ToNot(BeNil(), "created recorder with converted cassette file should not be nil")
		_ = rec.Stop()
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
	for k := range FuzzyRequestHeaders {
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
