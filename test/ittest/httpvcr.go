package ittest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
	"flag"
	"go.uber.org/fx"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"net/http"
	"strconv"
	"testing"
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
	RecorderOption  *HTTPVCROption
	RecorderMatcher cassette.MatcherFunc
}

// WithHttpPlayback enables remote HTTP server playback capabilities supported by `httpvcr`
// This mode requires apptest.Bootstrap to work
// Each top-level test should have corresponding recorded HTTP responses in `testdata` folder, or the test will fail.
// To enable record mode, use `go test ... -record` at CLI, or do it programmatically with HttpRecordingMode
// See https://github.com/cockroachdb/copyist for more details
func WithHttpPlayback(t *testing.T, opts ...HTTPVCROptions) test.Options {
	initial := HTTPVCROption{
		Name:           t.Name(),
		SavePath:       "testdata",
		RecordMatching: nil,
		Hooks: []RecorderHook{
			NewRecorderHook(FixedDurationHook(DefaultHTTPDuration), recorder.BeforeSaveHook),
			NewRecorderHook(InteractionIndexAwareHook(), recorder.BeforeSaveHook),
			NewRecorderHook(SanitizingHook(), recorder.BeforeSaveHook),
		},
		indexAwareWrapper: newIndexAwareMatcherWrapper(), // enforce order
	}

	var di RecorderDI
	testOpts := []test.Options{
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(
				httpRecorderProvider(initial, opts),
			),
			fx.Invoke(httpRecorderCleanup),
		),
		test.SubTestSetup(recorderDISetup(&di)),
		test.SubTestTeardown(recorderReset(&di)),
	}
	return test.WithOptions(testOpts...)
}

/****************************
	Functions
 ****************************/

// Client extract http.Client that provided by Recorder. If Recorder is not available, it returns nil
func Client(ctx context.Context) *http.Client {
	if rec, ok := ctx.Value(ckRecorder).(*recorder.Recorder); ok && rec != nil {
		return rec.GetDefaultClient()
	}
	return nil
}

// IsRecording returns true if HTTP VCR is in recording mode
func IsRecording(ctx context.Context) bool {
	if rec, ok := ctx.Value(ckRecorder).(*recorder.Recorder); ok && rec != nil {
		return rec.IsRecording()
	}
	return false
}

// AdditionalMatcherOptions temporarily add additional RecordMatcherOptions to the current test context
// on top of test's HTTPVCROptions.
// Note: The additional options take effect within the scope of sub-test. For test level options, use HttpRecordMatching
func AdditionalMatcherOptions(ctx context.Context, opts ...RecordMatcherOptions) {
	rec, ok := ctx.Value(ckRecorder).(*recorder.Recorder)
	if !ok || rec == nil {
		return
	}
	// merge matching options
	opt := ctx.Value(ckRecorderOption).(*HTTPVCROption)
	newOpts := make([]RecordMatcherOptions, len(opt.RecordMatching), len(opt.RecordMatching)+len(opts))
	copy(newOpts, opt.RecordMatching)
	newOpts = append(newOpts, opts...)

	// construct and set new matcher
	newMatcher := newCassetteMatcherFunc(newOpts, opt.indexAwareWrapper)
	rec.SetMatcher(newMatcher)
}

/*************************
	Options
 *************************/

// PackageHttpRecordingMode returns a suitetest.PackageOptions that enables HTTP recording mode for the entire package.
// This is usually used in TestMain function.
// Note: this option has no effect to tests using DisableHttpRecordingMode
// e.g.
// <code>
// 	func TestMain(m *testing.M) {
//		suitetest.RunTests(m,
//			PackageHttpRecordingMode(),
//		)
// 	}
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
// 				So use this mode on LOCAL DEV ONLY
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

// DisableHttpRecordOrdering disable HTTP interactions order matching.
// By default, HTTP interactions have to happen in the recorded order.
// When this option is used, HTTP interactions can happen in any order. However, each matched record can only replay once
func DisableHttpRecordOrdering() HTTPVCROptions {
	return func(opt *HTTPVCROption) {
		opt.indexAwareWrapper = nil
	}
}

/****************************
	Recorder Aware Context
 ****************************/

type recorderCtxKey struct{}
type optionCtxKey struct{}

var ckRecorder = recorderCtxKey{}
var ckRecorderOption = optionCtxKey{}

type recorderAwareContext struct {
	context.Context
	recorder   *recorder.Recorder
	origOption *HTTPVCROption
}

func contextWithRecorder(parent context.Context, rec *recorder.Recorder, opt *HTTPVCROption) *recorderAwareContext {
	return &recorderAwareContext{
		Context:    parent,
		recorder:   rec,
		origOption: opt,
	}
}

func (c *recorderAwareContext) Value(k interface{}) interface{} {
	switch k {
	case ckRecorder:
		return c.recorder
	case ckRecorderOption:
		return c.origOption
	default:
		return c.Context.Value(k)
	}
}

func recorderDISetup(di *RecorderDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		return contextWithRecorder(ctx, di.Recorder, di.RecorderOption), nil
	}
}

// recorderReset reset recorder to original state in case it changed
func recorderReset(di *RecorderDI) test.TeardownFunc {
	return func(ctx context.Context, t *testing.T) error {
		rec, ok := ctx.Value(ckRecorder).(*recorder.Recorder)
		if !ok {
			return nil
		}
		rec.SetMatcher(di.RecorderMatcher)
		return nil
	}
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
	Recorder             *recorder.Recorder
	CassetteMatcher      cassette.MatcherFunc
	HttpVCROption        *HTTPVCROption
	HttpClientCustomizer httpclient.ClientCustomizer `group:"http-client"`
}

func httpRecorderProvider(initial HTTPVCROption, opts []HTTPVCROptions) func(di vcrDI) (vcrOut, error) {
	return func(di vcrDI) (vcrOut, error) {
		opt := initial
		opts = append(opts, di.VCROptions...)
		for _, fn := range opts {
			fn(&opt)
		}
		rec, e := recorder.NewWithOptions(toRecorderOptions(opt))
		if e != nil {
			return vcrOut{}, e
		}

		// set matchers
		matcher := newCassetteMatcherFunc(opt.RecordMatching, opt.indexAwareWrapper)
		rec.SetMatcher(matcher)

		//set hooks
		order.SortStable(opt.Hooks, order.OrderedFirstCompare)
		for _, h := range opt.Hooks {
			rec.AddHook(h.Handler(), h.Kind())
		}

		return vcrOut{
			Recorder:        rec,
			CassetteMatcher: matcher,
			HttpVCROption:   &opt,
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

	name := opt.Name
	if len(opt.SavePath) != 0 {
		name = opt.SavePath + "/" + opt.Name + ".httpvcr"
	}
	return &recorder.Options{
		CassetteName:       name,
		Mode:               mode,
		RealTransport:      http.DefaultTransport,
		SkipRequestLatency: true,
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
			if e != errInteractionIDMismatch {
				logger.Debugf("HTTP interaction missing: %s - %v", record.Headers.Get(xInteractionIndexHeader), e)
			}
			return false
		}
		return true
	}
}
