package opensearchtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opensearch"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"github.com/cockroachdb/copyist"
	opensearchgo "github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"go.uber.org/fx"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"net/http"
	"testing"
	"time"
)

// IndexSuffix is the suffix we append to the index name when running opensearch tests, so that we don't
// corrupt the application's indices.
const IndexSuffix = "_test"

// IsRecording returns true if copyist is currently in recording mode.
// We wrap the copyist.IsRecording because we re-use the same commandline flag
// as the copyist library, and flag.Bool doesn't like it when you have two places
// that listen to the same flag
func IsRecording() bool {
	return copyist.IsRecording()
}

type Options func(opt *Option)
type Option struct {
	Name               string
	SavePath           string
	Mode               Mode
	RealTransport      http.RoundTripper
	SkipRequestLatency bool
	FuzzyJsonPaths     []string
	RecordDelay        time.Duration
}

// SetRecordDelay add delay between each request.
// Note: original request latency is applied by default. This is the additional delay between each requests
func SetRecordDelay(delay time.Duration) Options {
	return func(opt *Option) {
		SkipRequestLatency(false)(opt)
		opt.RecordDelay = delay
	}
}

func SetRecordMode(mode Mode) Options {
	return func(o *Option) {
		o.Mode = mode
	}
}

// SkipRequestLatency disable mimic request latency in playback mode. Has no effect during recording
// By default, original request latency during recording is applied in playback mode.
func SkipRequestLatency(skip bool) Options {
	return func(o *Option) {
		o.SkipRequestLatency = skip
	}
}

// ReplayMode override recording/playback mode. Default is ModeCommandline
func ReplayMode(mode Mode) Options {
	return func(o *Option) {
		o.Mode = mode
	}
}

// FuzzyJsonPaths ignore part of JSON body with JSONPath notation during playback mode.
// Useful for search queries with time-sensitive fields
// e.g. FuzzyJsonPaths("$.query.*.Time")
// JSONPath Syntax: https://goessner.net/articles/JsonPath/
func FuzzyJsonPaths(jsonPaths...string) Options {
	return func(o *Option) {
		o.FuzzyJsonPaths = append(o.FuzzyJsonPaths, jsonPaths...)
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
// *MatcherBodyModifiers to uber.FX.
func WithOpenSearchPlayback(options ...Options) test.Options {
	openSearchOption := Option {
		Mode: ModeCommandline,
	}
	for _, fn := range options {
		fn(&openSearchOption)
	}

	//var modifiers MatcherBodyModifiers
	//openSearchOption.RecordOptions = append(
	//	openSearchOption.RecordOptions,
	//	func(c *RecordOption) {
	//		c.Modifiers = &modifiers
	//	},
	//)

	var rec *recorder.Recorder
	testOpts := []test.Options{
		test.Setup(
			startRecording(&rec, options...),
		),
		apptest.WithFxOptions(
			fx.Decorate(func(c opensearchgo.Config) opensearchgo.Config {
				c.Transport = rec
				return c
			}),
			fx.Provide(
				IndexEditHookProvider(opensearch.FxGroup),
				//func() *MatcherBodyModifiers { return &MatcherBodyModifiers{} },
			),
		),
		test.Teardown(stopRecording()),
	}
	if openSearchOption.Mode == ModeRecording || openSearchOption.Mode == ModeCommandline && IsRecording(){
		testOpts = append(testOpts, apptest.WithFxOptions(
			fx.Provide(SearchDelayerHookProvider(opensearch.FxGroup, openSearchOption.RecordDelay)),
		))
	}
	return test.WithOptions(testOpts...)
}

func startRecording(recRef **recorder.Recorder, options ...Options) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		initial := func(c *Option) {
			c.Mode = ModeCommandline
			c.Name = t.Name()
			c.SavePath = "testdata"
		}
		opts := append([]Options{initial}, options...)
		var err error
		*recRef, err = NewRecorder(opts...)
		return contextWithRecorder(ctx, *recRef), err
	}
}

func stopRecording() test.TeardownFunc {
	return func(ctx context.Context, t *testing.T) error {
		if rec, ok := ctx.Value(ckRecorder{}).(*recorder.Recorder); ok {
			return rec.Stop()
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

func SearchDelayerHookProvider(group string, delay time.Duration) (fx.Annotated, fx.Annotated) {
	searchDelayer := SearchDelayerHook(delay)
	return fx.Annotated{
			Group: group, Target: func() opensearch.BeforeHook { return searchDelayer },
		},
		fx.Annotated{
			Group: group, Target: func() opensearch.AfterHook { return searchDelayer },
		}
}

type EditIndexForTestingHook struct {
	Suffix string
}

func (e *EditIndexForTestingHook) Order() int {
	return order.Highest
}

func NewEditingIndexForTestingHook() opensearch.BeforeHook {
	return &EditIndexForTestingHook{
		Suffix: IndexSuffix,
	}
}

func (e *EditIndexForTestingHook) Before(ctx context.Context, before opensearch.BeforeContext) context.Context {
	switch opt := before.Options.(type) {
	case *[]func(request *opensearchapi.SearchRequest):
		f := func(request *opensearchapi.SearchRequest) {
			var indices []string
			for _, index := range request.Index {
				indices = append(indices, index+e.Suffix)
			}
			request.Index = indices
		}
		*opt = append(*opt, f)
	case *[]func(request *opensearchapi.IndicesCreateRequest):
		f := func(request *opensearchapi.IndicesCreateRequest) {
			request.Index = request.Index + e.Suffix
		}
		*opt = append(*opt, f)
	case *[]func(request *opensearchapi.IndexRequest):
		f := func(request *opensearchapi.IndexRequest) {
			request.Index = request.Index + e.Suffix
		}
		*opt = append(*opt, f)
	case *[]func(request *opensearchapi.IndicesPutAliasRequest):
		f := func(request *opensearchapi.IndicesPutAliasRequest) {
			var indices []string
			for _, index := range request.Index {
				indices = append(indices, index+e.Suffix)
			}
			request.Index = indices
		}
		*opt = append(*opt, f)
	case *[]func(request *opensearchapi.IndicesDeleteAliasRequest):
		f := func(request *opensearchapi.IndicesDeleteAliasRequest) {
			var indices []string
			for _, index := range request.Index {
				indices = append(indices, index+e.Suffix)
			}
			request.Index = indices
		}
		*opt = append(*opt, f)
	case *[]func(request *opensearchapi.IndicesDeleteRequest):
		f := func(request *opensearchapi.IndicesDeleteRequest) {
			var indices []string
			for _, index := range request.Index {
				indices = append(indices, index+e.Suffix)
			}
			request.Index = indices
		}
		*opt = append(*opt, f)
	case *[]func(cfg *opensearchutil.BulkIndexerConfig):
		f := func(cfg *opensearchutil.BulkIndexerConfig) {
			cfg.Index = cfg.Index + e.Suffix
		}
		*opt = append(*opt, f)
	case *[]func(request *opensearchapi.SearchTemplateRequest):
		f := func(request *opensearchapi.SearchTemplateRequest) {
			var indices []string
			for _, index := range request.Index {
				indices = append(indices, index+e.Suffix)
			}
			request.Index = indices
		}
		*opt = append(*opt, f)
	}
	return ctx
}

func IndexEditHookProvider(group string) fx.Annotated {
	return fx.Annotated{
		Group:  group,
		Target: NewEditingIndexForTestingHook,
	}
}

/******************
	Context
 ******************/

type ckRecorder struct{}

type recorderContext struct {
	context.Context
	rec *recorder.Recorder
}

func (c recorderContext) Value(k interface{}) interface{} {
	switch k {
	case ckRecorder{}:
		return c.rec
	default:
		return c.Context.Value(k)
	}
}

func contextWithRecorder(parent context.Context, rec *recorder.Recorder) context.Context {
	return recorderContext{
		Context: parent,
		rec:     rec,
	}
}