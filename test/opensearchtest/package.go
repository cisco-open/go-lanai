package opensearchtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opensearch"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/httpvcr/recorder"
	"github.com/cockroachdb/copyist"
	opensearchgo "github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
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

type OpenSearchPlaybackOptions func(o *OpenSearchPlaybackOption)
type OpenSearchPlaybackOption struct {
	Mode          Mode
	RecordDelay   time.Duration
	RecordOptions []RecordOptions
}

func SetRecordDelay(duration time.Duration) OpenSearchPlaybackOptions {
	return func(o *OpenSearchPlaybackOption) {
		o.RecordDelay = duration
	}
}
func SetRecordMode(mode Mode) OpenSearchPlaybackOptions {
	return func(o *OpenSearchPlaybackOption) {
		o.Mode = mode
	}
}

// WithOpenSearchPlayback will setup the recorder, similar to crdb's copyist functionality
// where actual interactions with opensearch will be recorded, and then when the mode is set to
// ModeReplaying, the recorder will respond with its recorded responses.
// the parameter recordDelay defines how long of a delay is needed between a write to
// opensearch, and a read. opensearch does not immediately have writes available, so the only
// solution right now is to delay and reads that happen immediately after a write.
// For some reason, the refresh options on the index to opensearch are not working.
//
// To control what is being matched in the http vcr, this function will provide a
// *MatcherBodyModifierController to uber.FX.
func WithOpenSearchPlayback(options ...OpenSearchPlaybackOptions) test.Options {
	var openSearchOption OpenSearchPlaybackOption
	for _, fn := range options {
		fn(&openSearchOption)
	}

	var modifierController MatcherBodyModifierController
	openSearchOption.RecordOptions = append(
		openSearchOption.RecordOptions,
		func(c *RecordOption) {
			c.MatchBodyModifiers = modifierController.Modifier()
		},
	)

	determineMode(&openSearchOption.Mode)
	rec := recorder.Recorder{}
	testOpts := []test.Options{
		test.Setup(
			getRecording(
				&rec,
				openSearchOption.Mode,
				openSearchOption.RecordOptions...,
			),
		),
		apptest.WithFxOptions(
			fx.Decorate(func(c opensearchgo.Config) opensearchgo.Config {
				c.Transport = &rec
				return c
			}),
			fx.Provide(
				IndexEditHookProvider(opensearch.FxGroup, "test_"),
				func() *MatcherBodyModifierController { return &modifierController },
			),
		),
		test.Teardown(stopRecording(&rec)),
	}
	if openSearchOption.Mode == ModeRecording {
		testOpts = append(testOpts, apptest.WithFxOptions(
			fx.Provide(
				SearchDelayerHookProvider(
					opensearch.FxGroup,
					opensearch.FxGroup,
					openSearchOption.RecordDelay,
				),
			),
		))
	}
	return test.WithOptions(testOpts...)
}

func getRecording(rec *recorder.Recorder, mode Mode, options ...RecordOptions) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		r, err := GetRecorder(
			append(
				options,
				CassetteLocation(GetCassetteLocation()),
				ReplayMode(mode),
			)...,
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

func (s *SearchDelayer) Before(ctx context.Context, beforeContext opensearch.BeforeContext) context.Context {
	if beforeContext.CommandType() == opensearch.CmdSearch && s.lastEvent == opensearch.CmdIndex {
		time.Sleep(s.Delay)
	}
	return ctx
}

func (s *SearchDelayer) After(ctx context.Context, afterContext opensearch.AfterContext) context.Context {
	s.lastEvent = afterContext.CommandType()
	return ctx
}

func SearchDelayerHook(delay time.Duration) *SearchDelayer {
	return &SearchDelayer{Delay: delay}
}

func SearchDelayerHookProvider(beforeGroup string, afterGroup string, delay time.Duration) (fx.Annotated, fx.Annotated) {
	searchDelayer := SearchDelayerHook(delay)
	return fx.Annotated{
			Group: beforeGroup, Target: func() opensearch.BeforeHook { return searchDelayer },
		},
		fx.Annotated{
			Group: afterGroup, Target: func() opensearch.AfterHook { return searchDelayer },
		}
}

func EditIndexForTesting(prepend string) opensearch.BeforeHookFunc {
	return func(ctx context.Context, beforeContext opensearch.BeforeContext) context.Context {
		switch opt := beforeContext.Options.(type) {
		case *[]func(request *opensearchapi.SearchRequest):
			f := func(request *opensearchapi.SearchRequest) {
				var indices []string
				for _, index := range request.Index {
					indices = append(indices, prepend+index)
				}
				request.Index = indices
			}
			*opt = append(*opt, f)
		case *[]func(request *opensearchapi.IndicesCreateRequest):
			f := func(request *opensearchapi.IndicesCreateRequest) {
				request.Index = prepend + request.Index
			}
			*opt = append(*opt, f)
		case *[]func(request *opensearchapi.IndexRequest):
			f := func(request *opensearchapi.IndexRequest) {
				request.Index = prepend + request.Index
			}
			*opt = append(*opt, f)
		case *[]func(request *opensearchapi.IndicesDeleteRequest):
			f := func(request *opensearchapi.IndicesDeleteRequest) {
				var indices []string
				for _, index := range request.Index {
					indices = append(indices, prepend+index)
				}
				request.Index = indices
			}
			*opt = append(*opt, f)
		}
		return ctx
	}
}

func IndexEditHookProvider(group string, prepend string) fx.Annotated {
	return fx.Annotated{
		Group: group,
		Target: func() opensearch.BeforeHook {
			return opensearch.BeforeHookFunc(EditIndexForTesting(prepend))
		},
	}
}
