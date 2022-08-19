package opensearchtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opensearch"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/httpvcr/recorder"
	"github.com/cockroachdb/copyist"
	opensearchgo "github.com/opensearch-project/opensearch-go"
	"go.uber.org/fx"
	"testing"
	"time"
)

// IsRecording returns true if copyist is currently in recording mode.
// We wrap the copyist.IsRecording because we re-use the same commandline flag
// as the copyist library, and flag.Bool doesn't like it when you have two places
// that listen to the same flag
func IsRecording() bool {
	return copyist.IsRecording()
}

// determineMode will take the mode and determine what mode it should be depending on
// the commandline and environment variables
func determineMode(mode *Mode) {
	if *mode == ModeCommandline {
		// We let the commandline override this mode. Otherwise, this mode is determined
		// by the whatever it came in as
		if IsRecording() {
			*mode = ModeRecording
		} else {
			*mode = ModeReplaying
		}
	}
}

// WithOpenSearchPlayback will setup the recorder, similar to crdb's copyist functionality
// where actual interactions with opensearch will be recorded, and then when the mode is set to
// ModeReplaying, the recorder will respond with its recorded responses.
// the parameter recordDelay defines how long of a delay is needed between a write to
// opensearch, and a read. opensearch does not immediately have writes available, so the only
// solution right now is to delay and reads that happen immediately after a write.
// For some reason, the refresh options on the index to opensearch are not working.
func WithOpenSearchPlayback(mode Mode, recordDelay time.Duration) test.Options {
	determineMode(&mode)
	rec := recorder.Recorder{}
	testOpts := []test.Options{
		test.Setup(getRecording(&rec, mode)),
		apptest.WithFxOptions(
			fx.Decorate(func(c opensearchgo.Config) opensearchgo.Config {
				c.Transport = &rec
				return c
			}),
		),
		test.Teardown(stopRecording(&rec)),
	}
	if mode == ModeRecording {
		testOpts = append(testOpts, apptest.WithFxOptions(
			fx.Provide(
				SearchDelayerHookProvider(
					opensearch.FxOpenSearchHooksGroup,
					recordDelay,
				),
			),
		))
	}
	return test.WithOptions(testOpts...)
}

func getRecording(rec *recorder.Recorder, mode Mode) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		r, err := GetRecorder(
			CassetteLocation(GetCassetteLocation()),
			ReplayMode(mode),
		)
		if err != nil {
			return ctx, err
		}
		*rec = *r
		return ctx, nil
	}
}

func stopRecording(rec *recorder.Recorder) test.TeardownFunc {
	return func(ctx context.Context, t *testing.T) error {
		err := rec.Stop()
		if err != nil {
			return err
		}
		return nil
	}
}

// SearchDelayer will ensure that all searches that happen after inserting a document
// will have a delay so that the search can find all the documents.
type SearchDelayer struct {
	Delay time.Duration

	lastEvent opensearch.CommandType
}

func (s *SearchDelayer) BeforeHook() func(ctx opensearch.HookContext) {
	return func(ctx opensearch.HookContext) {
		if ctx.Cmd == opensearch.CmdSearch && s.lastEvent == opensearch.CmdIndex {
			time.Sleep(s.Delay)
		}
	}
}

func (s *SearchDelayer) AfterHook() func(ctx opensearch.HookContext) {
	return func(ctx opensearch.HookContext) {
		s.lastEvent = ctx.Cmd
	}
}

func SearchDelayerHook(delay time.Duration) opensearch.HookContainer {
	s := SearchDelayer{Delay: delay}
	return opensearch.HookContainer{
		Before: s.BeforeHook(),
		After:  s.AfterHook(),
	}
}

func SearchDelayerHookProvider(group string, delay time.Duration) fx.Annotated {
	return fx.Annotated{
		Group:  group,
		Target: func() opensearch.HookContainer { return SearchDelayerHook(delay) },
	}
}
