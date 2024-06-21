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
	"errors"
	"flag"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/integrate/httpclient"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/suitetest"
	"go.uber.org/fx"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func init() {
	// try register "record" flag, it may fail if it's already registered
	if flag.Lookup(CLIRecordModeFlag) == nil {
		flag.Bool(CLIRecordModeFlag, true, "record external interaction")
	}
}

type RecorderDI struct {
	fx.In
	Recorder        *recorder.Recorder
	RecorderOption  *recorder.Options
	RecorderMatcher cassette.MatcherFunc
	HTTPVCROption   *HTTPVCROption
}

type recorderDI struct {
	fx.In
	RecorderDI
	HTTPRecorder *HttpRecorder
}

// WithHttpPlayback enables remote HTTP server playback capabilities supported by `httpvcr`
// This mode requires apptest.Bootstrap to work
// Each top-level test should have corresponding recorded HTTP responses in `testdata` folder, or the test will fail.
// To enable record mode, use `go test ... --record-http` at CLI, or do it programmatically with HttpRecordingMode
// See https://github.com/cockroachdb/copyist for more details
func WithHttpPlayback(t *testing.T, opts ...HTTPVCROptions) test.Options {
	opts = append([]HTTPVCROptions{
		HttpRecordName(t.Name()),
		SanitizeHttpRecord(),
	}, opts...)

	var di recorderDI
	testOpts := []test.Options{
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(
				httpRecorderProvider(opts),
			),
			fx.Invoke(httpRecorderCleanup),
		),
		test.SubTestSetup(recorderDISetup(&di)),
		test.SubTestTeardown(recorderReset()),
	}
	return test.WithOptions(testOpts...)
}

/****************************
	Functions
 ****************************/

// Recorder extract HttpRecorder from given context. If HttpRecorder is not available, it returns nil
func Recorder(ctx context.Context) *HttpRecorder {
	if rec, ok := ctx.Value(ckRecorder).(*HttpRecorder); ok && rec.Recorder != nil {
		return rec
	}
	return nil
}

// Client extract http.Client that provided by Recorder. If Recorder is not available, it returns nil
func Client(ctx context.Context) *http.Client {
	if rec, ok := ctx.Value(ckRecorder).(*HttpRecorder); ok && rec.Recorder != nil {
		return rec.GetDefaultClient()
	}
	return nil
}

// IsRecording returns true if HTTP VCR is in recording mode
func IsRecording(ctx context.Context) bool {
	if rec, ok := ctx.Value(ckRecorder).(*HttpRecorder); ok && rec.Recorder != nil {
		return rec.IsRecording()
	}
	return false
}

// AdditionalMatcherOptions temporarily add additional RecordMatcherOptions to the current test context on top of test's HTTPVCROptions.
// Any changes made with this method can be reset via ResetRecorder. When using with WithHttpPlayback(), the reset is automatic per sub-test
// Note: The additional options take effect within the scope of sub-test. For test level options, use HttpRecordMatching.
func AdditionalMatcherOptions(ctx context.Context, opts ...RecordMatcherOptions) {
	rec, ok := ctx.Value(ckRecorder).(*HttpRecorder)
	if !ok || rec.Recorder == nil {
		return
	}
	// merge matching options
	newOpts := make([]RecordMatcherOptions, len(rec.Options.RecordMatching), len(rec.Options.RecordMatching)+len(opts))
	copy(newOpts, rec.Options.RecordMatching)
	newOpts = append(newOpts, opts...)

	// construct and set new matcher
	newMatcher := newCassetteMatcherFunc(newOpts, rec.Options.indexAwareWrapper)
	rec.SetMatcher(newMatcher)
}

// ResetRecorder revert the change made by AdditionalMatcherOptions.
func ResetRecorder(ctx context.Context) {
	rec, ok := ctx.Value(ckRecorder).(*HttpRecorder)
	if !ok || rec.Recorder == nil {
		return
	}
	rec.SetMatcher(rec.Matcher)
}

// StopRecorder stops the recorder extracted from the given context.
func StopRecorder(ctx context.Context) error {
	rec, ok := ctx.Value(ckRecorder).(*HttpRecorder)
	if !ok || rec.Recorder == nil {
		return fmt.Errorf("failed to stop recorder: no recorder found in context")
	}
	return rec.Stop()
}

/*************************
	Options
 *************************/

// PackageHttpRecordingMode returns a suitetest.PackageOptions that enables HTTP recording mode for the entire package.
// This is usually used in TestMain function.
// Note: this option has no effect to tests using DisableHttpRecordingMode
// e.g.
// <code>
//
//	func TestMain(m *testing.M) {
//		suitetest.RunTests(m,
//			PackageHttpRecordingMode(),
//		)
//	}
//
// </code>
func PackageHttpRecordingMode() suitetest.PackageOptions {
	return suitetest.Setup(func() error {
		return flag.Set(CLIRecordModeFlag, "true")
	})
}

// HttpRecordingMode returns a HTTPVCROptions that turns on Recording mode.
// Normally recording mode should be enabled via `go test` argument `-record-http`
// Note:	Record mode is forced off if flag is set to "-record-http=false" explicitly
// IMPORTANT:	When Record mode is enabled, all sub tests interact with actual HTTP remote service.
//
//	So use this mode on LOCAL DEV ONLY
func HttpRecordingMode() HTTPVCROptions {
	return func(opt *HTTPVCROption) {
		opt.Mode = ModeRecording
	}
}

// DisableHttpRecordingMode returns a HTTPVCROptions that force replaying mode regardless the command line flag
func DisableHttpRecordingMode() HTTPVCROptions {
	return func(opt *HTTPVCROption) {
		opt.Mode = ModeReplaying
	}
}

// HttpRecordName returns a HTTPVCROptions that set HTTP record's name
func HttpRecordName(name string) HTTPVCROptions {
	return func(opt *HTTPVCROption) {
		opt.Name = name
	}
}

// HttpRecordMatching returns a HTTPVCROptions that allows custom matching of recorded requests
func HttpRecordMatching(opts ...RecordMatcherOptions) HTTPVCROptions {
	return func(opt *HTTPVCROption) {
		opt.RecordMatching = append(opt.RecordMatching, opts...)
	}
}

// HttpRecorderHooks returns a HTTPVCROptions that adds recording hooks. If given hooks also implementing order.Ordered,
// the order will be respected
func HttpRecorderHooks(hooks ...RecorderHook) HTTPVCROptions {
	return func(opt *HTTPVCROption) {
		opt.Hooks = append(opt.Hooks, hooks...)
	}
}

// HttpRecordIgnoreHost convenient HTTPVCROptions that would ignore host when matching recorded requests,
// equivalent to HttpRecordMatching(IgnoreHost())
func HttpRecordIgnoreHost() HTTPVCROptions {
	return HttpRecordMatching(IgnoreHost())
}

// HttpRecordOrdering toggles HTTP interactions order matching.
// When enforced, HTTP interactions have to happen in the recorded order.
// Otherwise, HTTP interactions can happen in any order, but each matched record can only replay once
// By default, record ordering is enabled
func HttpRecordOrdering(enforced bool) HTTPVCROptions {
	return func(opt *HTTPVCROption) {
		if enforced && opt.indexAwareWrapper == nil {
			opt.indexAwareWrapper = newIndexAwareMatcherWrapper()
		} else if !enforced {
			opt.indexAwareWrapper = nil
		}
	}
}

// DisableHttpRecordOrdering disable HTTP interactions order matching.
// By default, HTTP interactions have to happen in the recorded order.
// When this option is used, HTTP interactions can happen in any order. However, each matched record can only replay once
func DisableHttpRecordOrdering() HTTPVCROptions {
	return HttpRecordOrdering(false)
}

// HttpTransport override the RealTransport during recording mode. This option has no effect in playback mode
func HttpTransport(transport http.RoundTripper) HTTPVCROptions {
	return func(opt *HTTPVCROption) {
		opt.RealTransport = transport
	}
}

// ApplyHttpLatency apply recorded HTTP latency. By default, HTTP latency is not applied for faster test run. This option has no effect in recording mode
func ApplyHttpLatency() HTTPVCROptions {
	return func(opt *HTTPVCROption) {
		opt.SkipRequestLatency = false
	}
}

// SanitizeHttpRecord install a hook to sanitize request and response before they are saved in file.
// See SanitizingHook for details.
func SanitizeHttpRecord() HTTPVCROptions {
	return func(opt *HTTPVCROption) {
		opt.Hooks = append(opt.Hooks, NewRecorderHook(SanitizingHook(), recorder.BeforeSaveHook))
	}
}

// FixedHttpRecordDuration install a hook to set a fixed duration on interactions before they are saved.
// Otherwise, the actual latency will be recorded.
// When HTTPVCROption.SkipRequestLatency is set to false, the recorded duration will be applied during playback
// See FixedDurationHook for details.
func FixedHttpRecordDuration(duration time.Duration) HTTPVCROptions {
	return func(opt *HTTPVCROption) {
		opt.Hooks = append(opt.Hooks, NewRecorderHook(FixedDurationHook(duration), recorder.BeforeSaveHook))
	}
}

/****************************
	RawRecorder Aware Context
 ****************************/

type recorderCtxKey struct{}

var ckRecorder = recorderCtxKey{}

type recorderAwareContext struct {
	context.Context
	recorder *HttpRecorder
}

func contextWithRecorder(parent context.Context, rec *HttpRecorder) context.Context {
	if rec == nil {
		return parent
	}
	return &recorderAwareContext{
		Context:  parent,
		recorder: rec,
	}
}

func (c *recorderAwareContext) Value(k interface{}) interface{} {
	switch k {
	case ckRecorder:
		return c.recorder
	default:
		return c.Context.Value(k)
	}
}

func recorderDISetup(di *recorderDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		return contextWithRecorder(ctx, di.HTTPRecorder), nil
	}
}

// recorderReset automatically reset recorder to original state in case it changed
func recorderReset() test.TeardownFunc {
	return func(ctx context.Context, t *testing.T) error {
		ResetRecorder(ctx)
		return nil
	}
}

/*************************
	HttpRecorder
 *************************/

// HttpRecorder wrapper of recorder.RawRecorder, used to hold some value that normally inaccessible via wrapped recorder.RawRecorder.
// Note: This type is for other test utilities to re-configure recorder.RawRecorder
type HttpRecorder struct {
	*recorder.Recorder
	RawOptions *recorder.Options
	Matcher    cassette.MatcherFunc
	Options    *HTTPVCROption
}

// ContextWithNewHttpRecorder is a convenient function that create a new HTTP recorder and store it in context.
// The returned context can be used with context value accessor such as Client(ctx), IsRecording(ctx), AdditionalMatcherOptions(ctx), etc.
// See NewHttpRecorder
func ContextWithNewHttpRecorder(ctx context.Context, opts ...HTTPVCROptions) (context.Context, error) {
	rec, e := NewHttpRecorder(opts...)
	if e != nil {
		return nil, e
	}
	return contextWithRecorder(ctx, rec), nil
}

// NewHttpRecorder create a new HttpRecorder. Commonly used by:
// - other test utilities that relies on http recording. (e.g. opensearchtest, consultest, etc.)
// - unit tests that doesn't bootstrap dependency injection
func NewHttpRecorder(opts ...HTTPVCROptions) (*HttpRecorder, error) {
	opt := HTTPVCROption{
		SavePath: "testdata",
		Hooks: []RecorderHook{
			NewRecorderHook(InteractionIndexAwareHook(), recorder.BeforeSaveHook),
		},
		SkipRequestLatency: true,
		indexAwareWrapper:  newIndexAwareMatcherWrapper(), // enforce order
	}
	for _, fn := range opts {
		fn(&opt)
	}
	rawOpts := toRecorderOptions(opt)
	rec, e := recorder.NewWithOptions(rawOpts)
	if e != nil {
		return nil, e
	}

	// set matchers
	matcher := newCassetteMatcherFunc(opt.RecordMatching, opt.indexAwareWrapper)
	rec.SetMatcher(matcher)

	//set hooks
	order.SortStable(opt.Hooks, order.OrderedFirstCompare)
	for _, h := range opt.Hooks {
		rec.AddHook(h.Handler(), h.Kind())
	}
	return &HttpRecorder{
		Recorder:   rec,
		RawOptions: rawOpts,
		Matcher:    matcher,
		Options:    &opt,
	}, nil
}

/*************************
	Internals
 *************************/

type vcrDI struct {
	fx.In
	VCROptions []HTTPVCROptions `group:"http-vcr"`
}

type vcrOut struct {
	fx.Out
	HTTPRecorder         *HttpRecorder
	RawRecorder          *recorder.Recorder
	CassetteMatcher      cassette.MatcherFunc
	HttpVCROption        *HTTPVCROption
	RawRecorderOption    *recorder.Options
	HttpClientCustomizer httpclient.ClientCustomizer `group:"http-client"`
}

func httpRecorderProvider(opts []HTTPVCROptions) func(di vcrDI) (vcrOut, error) {
	return func(di vcrDI) (vcrOut, error) {
		finalOpts := append(opts, di.VCROptions...)
		rec, e := NewHttpRecorder(finalOpts...)
		if e != nil {
			return vcrOut{}, e
		}
		return vcrOut{
			HTTPRecorder:      rec,
			RawRecorder:       rec.Recorder,
			CassetteMatcher:   rec.Matcher,
			HttpVCROption:     rec.Options,
			RawRecorderOption: rec.RawOptions,
			HttpClientCustomizer: httpclient.ClientCustomizerFunc(func(opt *httpclient.ClientOption) {
				opt.HTTPClient = rec.GetDefaultClient()
			}),
		}, nil
	}
}

func findBoolFlag(name string) (ret *bool) {
	flag.Visit(func(f *flag.Flag) {
		if f.Name != name {
			return
		}
		var b bool
		b, e := strconv.ParseBool(f.Value.String())
		if e != nil {
			b = true // default to true
		}
		ret = &b
	})
	return
}

func toRecorderOptions(opt HTTPVCROption) *recorder.Options {
	cliFlag := findBoolFlag(CLIRecordModeFlag)
	mode := recorder.ModeReplayOnly
	switch opt.Mode {
	case ModeRecording:
		if cliFlag == nil || *cliFlag {
			mode = recorder.ModeRecordOnly
		}
	case ModeCommandline:
		if cliFlag != nil && *cliFlag {
			mode = recorder.ModeRecordOnly
		}
	default:
	}

	name := opt.Name + ".httpvcr"
	if len(opt.SavePath) != 0 {
		name = opt.SavePath + "/" + opt.Name + ".httpvcr"
	}
	return &recorder.Options{
		CassetteName:       name,
		Mode:               mode,
		RealTransport:      opt.RealTransport,
		SkipRequestLatency: opt.SkipRequestLatency,
	}
}

func newCassetteMatcherFunc(opts []RecordMatcherOptions, indexAwareMatcher *indexAwareMatcherWrapper) cassette.MatcherFunc {
	matcherFn := NewRecordMatcher(opts...)
	if indexAwareMatcher == nil {
		return wrapRecordRequestMatcher(matcherFn)
	}
	return wrapRecordRequestMatcher(indexAwareMatcher.MatcherFunc(RecordMatcherFunc(matcherFn)))
}

func httpRecorderCleanup(lc fx.Lifecycle, rec *recorder.Recorder) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return rec.Stop()
		},
	})
}

func wrapRecordRequestMatcher(fn GenericMatcherFunc[*http.Request, cassette.Request]) cassette.MatcherFunc {
	return func(out *http.Request, record cassette.Request) bool {
		if e := fn(out, record); e != nil {
			if !errors.Is(e, errInteractionIDMismatch) {
				logger.Debugf("HTTP interaction missing: %s - %v: expect %s, but got %s",
					record.Headers.Get(xInteractionIndexHeader), e,
					fmt.Sprintf(`%s "%s"`, record.Method, record.URL),
					fmt.Sprintf(`%s "%s"`, out.Method, out.URL.String()))
			}
			return false
		}
		return true
	}
}
