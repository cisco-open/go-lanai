package ittest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"github.com/cockroachdb/copyist"
	"go.uber.org/fx"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"net/http"
	"testing"
)

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

	testOpts := []test.Options{
		apptest.WithFxOptions(
			fx.Provide(
				httpRecorderProvider(initial, opts),
			),
			fx.Invoke(httpRecorderCleanup),
		),
	}
	return test.WithOptions(testOpts...)
}

// IsRecording returns true if copyist is in recording mode
func IsRecording() bool {
	return copyist.IsRecording()
}

/*************************
	Options
 *************************/

// EnableHttpRecordMode returns a HttpVCROptions that enable Recording mode
// Normally recording mode should be enabled via `go test` argument `-record`
// IMPORTANT: When Record mode is enabled, all tests interact with actual HTTP remote service.
// 			  So use this mode on LOCAL DEV ONLY
func EnableHttpRecordMode() HttpVCROptions {
	return func(opt *HttpVCROption) {
		opt.Mode = ModeRecording
	}
}

// HttpRecordName returns a HttpVCROptions that set HTTP record's name
func HttpRecordName(name string) HttpVCROptions {
	return func(opt *HttpVCROption) {
		opt.Name = name
	}
}

// HttpRecordCustomMatching returns a HttpVCROptions that allows custom matching of recorded requests
func HttpRecordCustomMatching(opts...RecordMatcherOptions) HttpVCROptions {
	return func(opt *HttpVCROption) {
		opt.RecordMatching = append(opt.RecordMatching, opts...)
	}
}

// HttpRecordIgnoreHost returns a HttpVCROptions that would ignore host when matching recorded requests
func HttpRecordIgnoreHost() HttpVCROptions {
	return HttpRecordCustomMatching(func(opt *RecordMatcherOption) {
		opt.URLMatcher = RecordURLMatcherFunc(NewRecordHostIgnoringURLMatcher())
	})
}



/*************************
	Internals
 *************************/

func toRecorderOptions(opt HttpVCROption) *recorder.Options {
	mode := recorder.ModeReplayOnly
	switch opt.Mode {
	case ModeRecording:
		mode = recorder.ModeRecordOnly
	case ModeCommandline:
		// TODO
	}
	name := opt.Name
	if len(opt.SavePath) != 0 {
		name = opt.SavePath + "/" + opt.Name
	}
	return &recorder.Options{
		CassetteName:       name,
		Mode:               mode,
		RealTransport:      http.DefaultTransport,
		SkipRequestLatency: true,
	}
}

func httpRecorderProvider(initial HttpVCROption, opts []HttpVCROptions) func() (*recorder.Recorder, error) {
	return func() (*recorder.Recorder, error) {
		for _, fn := range opts {
			fn(&initial)
		}
		rec, e := recorder.NewWithOptions(toRecorderOptions(initial))
		if e != nil {
			return nil, e
		}
		matchFn := NewRecordMatcher(initial.RecordMatching...)
		rec.SetMatcher(wrapRecordRequestMatcher(matchFn))
		return rec, nil
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

