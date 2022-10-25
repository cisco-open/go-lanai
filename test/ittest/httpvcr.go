package ittest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
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

const CLIRecordModeFlag = "record"

func init() {
	// try register "record" flag, it may fail if it's already registered
	if flag.Lookup(CLIRecordModeFlag) == nil {
		flag.Bool(CLIRecordModeFlag, true, "record external interaction")
	}
}

type RecorderDI struct {
	fx.In
	Recorder *recorder.Recorder
}

// WithHttpPlayback enables remote HTTP server playback capabilities supported by `httpvcr`
// This mode requires apptest.Bootstrap to work
// Each top-level test should have corresponding recorded HTTP responses in `testdata` folder, or the test will fail.
// To enable record mode, use `go test ... -record` at CLI, or do it programmatically with EnableHttpRecordMode
// See https://github.com/cockroachdb/copyist for more details
func WithHttpPlayback(t *testing.T, opts ...HttpVCROptions) test.Options {
	initial := HttpVCROption{
		Name:     t.Name(),
		SavePath: "testdata",
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
	}
	return test.WithOptions(testOpts...)
}

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

/*************************
	Options
 *************************/

// EnablePackageHttpRecordMode returns a suitetest.PackageOptions that enables HTTP recording mode for the entire package.
// This is usually used in TestMain function.
// Note: this option has no effect to tests using DisableHttpRecordMode
// e.g.
// <code>
// 	func TestMain(m *testing.M) {
//		suitetest.RunTests(m,
//			EnablePackageHttpRecordMode(),
//		)
// 	}
// </code>
func EnablePackageHttpRecordMode() suitetest.PackageOptions {
	return suitetest.Setup(func() error {
		return flag.Set(CLIRecordModeFlag, "true")
	})
}

// EnableHttpRecordMode returns a HttpVCROptions that enable Recording mode.
// Normally recording mode should be enabled via `go test` argument `-record`
// Note: 	  Record mode is forced off if flag is set to "-record=false" explicitly
// IMPORTANT: When Record mode is enabled, all tests interact with actual HTTP remote service.
// 			  So use this mode on LOCAL DEV ONLY
func EnableHttpRecordMode() HttpVCROptions {
	return func(opt *HttpVCROption) {
		opt.Mode = ModeRecording
	}
}

// DisableHttpRecordMode returns a HttpVCROptions that force replaying mode regardless the command line flag
func DisableHttpRecordMode() HttpVCROptions {
	return func(opt *HttpVCROption) {
		opt.Mode = ModeReplaying
	}
}

// HttpRecordName returns a HttpVCROptions that set HTTP record's name
func HttpRecordName(name string) HttpVCROptions {
	return func(opt *HttpVCROption) {
		opt.Name = name
	}
}

// HttpRecordCustomMatching returns a HttpVCROptions that allows custom matching of recorded requests
func HttpRecordCustomMatching(opts ...RecordMatcherOptions) HttpVCROptions {
	return func(opt *HttpVCROption) {
		opt.RecordMatching = append(opt.RecordMatching, opts...)
	}
}

// HttpRecordIgnoreHost returns a HttpVCROptions that would ignore host when matching recorded requests
func HttpRecordIgnoreHost() HttpVCROptions {
	matching := HttpRecordCustomMatching(func(opt *RecordMatcherOption) {
		opt.URLMatcher = RecordURLMatcherFunc(NewRecordHostIgnoringURLMatcher())
	})
	hook := recorder.Hook{
		Handler: HostIgnoringHook(),
		Kind:    recorder.BeforeSaveHook,
	}
	return func(opt *HttpVCROption) {
		matching(opt)
		opt.Hooks = append(opt.Hooks, hook)
	}
}

/****************************
	Recorder Aware Context
 ****************************/

type recorderContextKey struct{}

var ckRecorder = recorderContextKey{}
var ckOrigMatcher = recorderContextKey{}

type recorderAwareContext struct {
	context.Context
	recorder    *recorder.Recorder
	origMatcher cassette.MatcherFunc
}

func contextWithRecorder(parent context.Context, rec *recorder.Recorder) *recorderAwareContext {
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

func recorderDISetup(di *RecorderDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		return contextWithRecorder(ctx, di.Recorder), nil
	}
}

func recorderReset(di *RecorderDI) test.TeardownFunc {
	// TODO
	return func(ctx context.Context, t *testing.T) error {
		rec, ok := ctx.Value(ckRecorder).(*recorder.Recorder)
		if !ok {
			return nil
		}
		rec.IsRecording()
		return nil
	}
}

/*************************
	Internals
 *************************/

type vcrOut struct {
	fx.Out
	Recorder             *recorder.Recorder
	HttpClientCustomizer httpclient.ClientCustomizer `group:"http-client"`
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

func toRecorderOptions(opt HttpVCROption) *recorder.Options {
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

func httpRecorderProvider(initial HttpVCROption, opts []HttpVCROptions) func() (vcrOut, error) {
	return func() (vcrOut, error) {
		for _, fn := range opts {
			fn(&initial)
		}
		rec, e := recorder.NewWithOptions(toRecorderOptions(initial))
		if e != nil {
			return vcrOut{}, e
		}

		// set matchers
		matchFn := NewRecordMatcher(initial.RecordMatching...)
		idMatchFn := NewRecordIndexAwareMatcher()
		rec.SetMatcher(wrapRecordRequestMatcher(AndMatcher(matchFn, idMatchFn)))

		//set hooks
		for _, h := range initial.Hooks {
			rec.AddHook(h.Handler, h.Kind)
		}
		rec.AddHook(FixedDurationHook(DefaultHttpDuration), recorder.BeforeSaveHook)
		rec.AddHook(InteractionIndexAwareHook(), recorder.BeforeSaveHook)
		rec.AddHook(SanitizingHook(), recorder.BeforeSaveHook)

		return vcrOut{
			Recorder: rec,
			HttpClientCustomizer: RecordingHttpClientCustomizer{Recorder: rec},
		}, nil
	}
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
			logger.Debugf("HTTP interaction missing: %v", e)
			return false
		}
		return true
	}
}
