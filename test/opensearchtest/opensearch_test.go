package opensearchtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opensearch"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing/instrument"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"github.com/onsi/gomega"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/opentracing/opentracing-go/mocktracer"
	"go.uber.org/fx"
	"net/http"
	"testing"
	"time"
)

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

type FakeService struct {
	Repo opensearch.Repo[GenericAuditEvent]
}

type fakeServiceDI struct {
	fx.In
	Client opensearch.OpenClient
}

func NewFakeService(di fakeServiceDI) FakeService {
	return FakeService{
		Repo: opensearch.NewRepo(&GenericAuditEvent{}, di.Client),
	}
}

type opensearchDI struct {
	fx.In
	FakeService            FakeService
	Properties    *opensearch.Properties
	BodyModifiers *MatcherBodyModifiers
}

func TestScopeController(t *testing.T) {
	di := &opensearchDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithOpenSearchPlayback(
			SetRecordMode(ModeCommandline),
			SetRecordDelay(time.Millisecond*1500),
		),
		apptest.WithTimeout(time.Minute),
		apptest.WithModules(opensearch.Module),
		apptest.WithFxOptions(
			fx.Provide(NewFakeService),
		),
		apptest.WithProperties(
			"data.logging.level: debug",
			"log.levels.data: debug",
		),
		apptest.WithDI(di),
		test.SubTestSetup(SetupOpenSearchTest(di)),
		test.GomegaSubTest(SubTestRecording(di), "SubTestRecording"),
		test.GomegaSubTest(SubTestHooks(di), "SubTestHooks"),
		test.GomegaSubTest(SubTestTracer(di), "SubTestTracer"),
		test.GomegaSubTest(SubTestPing(di), "SubTestPing"),
		test.GomegaSubTest(SubTestTimeBasedQuery(di), "SubTestTimeBasedQuery"),
	)
}

func SetupOpenSearchTest(di *opensearchDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		return SetupPrepareOpenSearchData(
			ctx,
			di.FakeService.Repo,
			time.Date(2020, 0, 0, 0, 0, 0, 0, time.UTC),
			time.Date(2022, 0, 0, 0, 0, 0, 0, time.UTC),
		)
	}
}

func SubTestRecording(di *opensearchDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []map[string]interface{}{
						{
							"match": map[string]interface{}{
								"SubType": "SYNCHRONIZED",
							},
						},
					},
				},
			},
		}

		var dest []GenericAuditEvent
		_, err := di.FakeService.Repo.Search(context.Background(), &dest, query,
			opensearch.Search.WithIndex("auditlog"),
			opensearch.Search.WithRequestCache(false),
		)
		if err != nil {
			t.Fatalf("unable to search for document")
		}
		g.Expect(len(dest)).To(gomega.Equal(2))
		// These values come from inspecting what was randomly generated. They were then
		// manually recorded and placed here. The random values will be produced
		// as the same values everytime because the random seed will stay const
		g.Expect(dest[0].Client_ID).To(gomega.Equal("ibdei"))
		g.Expect(dest[0].Orig_User).To(gomega.Equal("iagjb"))
		g.Expect(dest[1].Client_ID).To(gomega.Equal("eejbi"))
		g.Expect(dest[1].Orig_User).To(gomega.Equal("edcdi"))

		testEvent := GenericAuditEvent{
			Client_ID: "TESTING TESTING",
			SubType:   "SYNCHRONIZED",
			Time:      time.Date(2019, 10, 15, 0, 0, 0, 0, time.UTC),
		}

		err = di.FakeService.Repo.Index(context.Background(), "auditlog", testEvent)
		if err != nil {
			t.Fatalf("unable to create document in index: %v", err)
		}
		totalHits, err := di.FakeService.Repo.Search(context.Background(), &dest, query,
			opensearch.Search.WithIndex("auditlog"),
		)
		if err != nil {
			t.Fatalf("unable to search for document: %v", err)
		}
		g.Expect(totalHits).To(gomega.Equal(3))
		g.Expect(len(dest)).To(gomega.Equal(3))
		g.Expect(dest[2].Client_ID).To(gomega.Equal(testEvent.Client_ID))
	}
}

// SubTestHooks tests to ensure that the hooks are called in their proper before/after order.
// The test will only use the repo.Search method to trigger the hooks.
func SubTestHooks(di *opensearchDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		searchQuery := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []map[string]interface{}{
						{
							"match": map[string]interface{}{
								"SubType": "SYNCHRONIZED",
							},
						},
					},
				},
			},
		}

		errorQuery := map[string]interface{}{
			"query": map[string]interface{}{
				"boolmispelled": map[string]interface{}{
					"must": []map[string]interface{}{
						{
							"match": map[string]interface{}{
								"SubType": "SYNCHRONIZED",
							},
						},
					},
				},
			},
		}

		type fields struct {
			beforeHooks []opensearch.BeforeHook
			afterHooks  []opensearch.AfterHook
		}
		type args struct {
			ctx     context.Context
			dest    []GenericAuditEvent
			query   map[string]interface{}
			options []opensearch.Option[opensearchapi.SearchRequest]
		}
		tests := []struct {
			name   string
			fields fields
			args   args
		}{
			{
				name: "Test Before hook gets called before the After hook, and shares context",
				fields: fields{
					beforeHooks: []opensearch.BeforeHook{
						opensearch.BeforeHookBase{
							Identifier: "key adder",
							F: func(bctx context.Context, before opensearch.BeforeContext) context.Context {
								bctx = context.WithValue(bctx, "key1", "val1")
								return bctx
							},
						},
					},
					afterHooks: []opensearch.AfterHook{
						opensearch.AfterHookBase{
							Identifier: "key checker",
							F: func(actx context.Context, after opensearch.AfterContext) context.Context {
								val := actx.Value("key1").(string)
								if val != "val1" {
									t.Error("Before hook was not called before this After hook")
								}
								return actx
							},
						},
					},
				},
				args: args{
					ctx: context.Background(), dest: []GenericAuditEvent{}, query: searchQuery,
					options: []opensearch.Option[opensearchapi.SearchRequest]{
						opensearch.Search.WithIndex("auditlog"),
					},
				},
			},
			{
				name: "Test After hook can detect error and manipulate response",
				fields: fields{
					afterHooks: []opensearch.AfterHook{
						opensearch.AfterHookBase{
							Identifier: "error checker",
							F: func(ctx context.Context, after opensearch.AfterContext) context.Context {
								// we check for nil in response because if we are not in recording mode, we may have an empty resp
								if after.Resp != nil {
									if !after.Resp.IsError() {
										t.Error("expected an error, we got nil")
									} else {
										// good, but we need the test to pass and want to test
										// out being able to manipulate responses
										after.Resp.StatusCode = http.StatusOK
									}
								}
								return ctx
							},
						},
					},
				},
				args: args{
					ctx: context.Background(), dest: []GenericAuditEvent{}, query: errorQuery,
					options: []opensearch.Option[opensearchapi.SearchRequest]{
						opensearch.Search.WithIndex("auditlog"),
					},
				},
			},
			{
				name: "Test adding an option to change to an empty index to cause read failure",
				fields: fields{
					beforeHooks: []opensearch.BeforeHook{
						opensearch.BeforeHookBase{
							Identifier: "index modifier",
							F: func(ctx context.Context, before opensearch.BeforeContext) context.Context {
								// change all the request indices to be from indices that don't exist
								switch opt := before.Options.(type) {
								case *[]func(request *opensearchapi.SearchRequest):
									f := func(request *opensearchapi.SearchRequest) {
										var indices []string
										for _, _ = range request.Index {
											indices = append(indices, "dontexist")
										}
										request.Index = indices
									}
									*opt = append(*opt, f)
								default:
									t.Errorf("these tests should only run on search requests")
								}
								return ctx
							},
						},
					},
					afterHooks: []opensearch.AfterHook{
						opensearch.AfterHookBase{
							Identifier: "error checker",
							F: func(ctx context.Context, after opensearch.AfterContext) context.Context {
								if after.Resp != nil {
									if !after.Resp.IsError() {
										t.Error("expected an error, we got nil")
									} else {
										// good, but we need the test to pass and want to test
										// out being able to manipulate responses
										after.Resp.StatusCode = http.StatusOK
									}
								} else {
									t.Error("expected response to not be nil")
								}
								return ctx
							},
						},
					},
				},
				args: args{
					ctx: context.Background(), dest: []GenericAuditEvent{}, query: searchQuery,
					options: []opensearch.Option[opensearchapi.SearchRequest]{
						opensearch.Search.WithIndex("auditlog"),
					},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				for _, hook := range tt.fields.beforeHooks {
					di.FakeService.Repo.AddBeforeHook(hook)
				}
				for _, hook := range tt.fields.afterHooks {
					di.FakeService.Repo.AddAfterHook(hook)
				}
				_, err := di.FakeService.Repo.Search(
					tt.args.ctx,
					&tt.args.dest,
					tt.args.query,
					tt.args.options...,
				)
				if err != nil {
					t.Fatalf("unable to search: %v", err)
				}
				for _, hook := range tt.fields.beforeHooks {
					di.FakeService.Repo.RemoveBeforeHook(hook)
				}
				for _, hook := range tt.fields.afterHooks {
					di.FakeService.Repo.RemoveAfterHook(hook)
				}
			})
		}
	}
}

func SubTestTracer(di *opensearchDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		searchQuery := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []map[string]interface{}{
						{
							"match": map[string]interface{}{
								"SubType": "SYNCHRONIZED",
							},
						},
					},
				},
			},
		}
		errorQuery := map[string]interface{}{
			"query": map[string]interface{}{
				"boolmispelled": map[string]interface{}{
					"must": []map[string]interface{}{
						{
							"match": map[string]interface{}{
								"SubType": "SYNCHRONIZED",
							},
						},
					},
				},
			},
		}
		type args struct {
			ctx       context.Context
			dest      []GenericAuditEvent
			query     map[string]interface{}
			options   []opensearch.Option[opensearchapi.SearchRequest]
			validator func(*testing.T, []*mocktracer.MockSpan)
		}
		tests := []struct {
			name string
			args args
		}{
			{
				name: "Test normal trace",
				args: args{
					ctx: context.Background(), dest: []GenericAuditEvent{}, query: searchQuery,
					options: []opensearch.Option[opensearchapi.SearchRequest]{
						opensearch.Search.WithIndex("auditlog"),
					},
					validator: func(t *testing.T, spans []*mocktracer.MockSpan) {
						if len(spans) != 1 {
							t.Errorf("expected length of span to be equal to 1, got: %v", len(spans))
							return
						}
						command := spans[0].Tag("command")
						if command.(opensearch.CommandType) != opensearch.CmdSearch {
							t.Errorf("expected command to be :%v, got: %v", opensearch.CmdSearch, command)
						}

						hits := spans[0].Tag("hits")
						expectedHits := 3
						if hits.(int) != expectedHits {
							t.Errorf("expected hits: %v, got: %v", expectedHits, hits.(int))
						}
					},
				},
			},
			{
				name: "Test error in query",
				args: args{
					ctx: context.Background(), dest: []GenericAuditEvent{}, query: errorQuery,
					options: []opensearch.Option[opensearchapi.SearchRequest]{
						opensearch.Search.WithIndex("auditlog"),
					},
					validator: func(t *testing.T, spans []*mocktracer.MockSpan) {
						if len(spans) != 1 {
							t.Errorf("expected length of span to be equal to 1, got: %v", len(spans))
							return
						}
						command := spans[0].Tag("command")
						if command.(opensearch.CommandType) != opensearch.CmdSearch {
							t.Errorf("expected command to be :%v, got: %v", opensearch.CmdSearch, command)
						}

						hits := spans[0].Tag("hits")
						if hits != nil {
							t.Errorf("expected nil hits")
						}

						statusCode := spans[0].Tag("status code")
						if statusCode.(int) != http.StatusBadRequest {
							t.Errorf("expected status code: %v, got: %v", http.StatusBadRequest, statusCode)
						}

					},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tracer := mocktracer.New()
				openTracer := instrument.OpenSearchTracerHook(tracer)
				di.FakeService.Repo.AddBeforeHook(openTracer)
				di.FakeService.Repo.AddAfterHook(openTracer)

				op := tracing.WithTracer(tracer)
				tt.args.ctx = op.NewSpanOrDescendant(tt.args.ctx)

				// ignore errors, we only care about the trace
				_, _ = di.FakeService.Repo.Search(
					tt.args.ctx,
					&tt.args.dest,
					tt.args.query,
					tt.args.options...,
				)
				spans := tracer.FinishedSpans()
				tt.args.validator(t, spans)
				di.FakeService.Repo.RemoveBeforeHook(openTracer)
				di.FakeService.Repo.RemoveAfterHook(openTracer)
			})
		}
	}
}

func SubTestPing(di *opensearchDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		err := di.FakeService.Repo.Ping(ctx)
		if err != nil {
			t.Fatalf("unable to search for document")
		}
	}
}

// SubTestTimeBasedQuery will test that the MatcherBodyModifier can be used to ignore
// a portion of the request that is used to compare requests in the httpvcr/recorder.Recorder
func SubTestTimeBasedQuery(di *opensearchDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		defer di.BodyModifiers.Clear()
		di.BodyModifiers.Append(IgnoreGJSONPaths(t,
			"query.bool.filter.range.Time",
		))

		var TimeQuery map[string]interface{}
		// If we are recording, we want to use a valid time query, however to test
		// that the time portion is ignored with our `IgnoreFilterRangeTime` function
		// we change the time query to something that doesn't make sense to check if
		// the test that the search (and httpvcr underneath) still matches a request
		if IsRecording() {
			TimeQuery = map[string]interface{}{
				"gt": "2020-01-10||/M",
				"lt": "2021-01-10||/M",
			}
		} else {
			TimeQuery = map[string]interface{}{
				"gt": "2000-01-10||/M",
				"lt": "2001-01-10||/M",
			}
		}
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"filter": map[string]interface{}{
						"range": map[string]interface{}{
							"Time": TimeQuery,
						},
					},
				},
			},
		}
		var dest []GenericAuditEvent
		_, err := di.FakeService.Repo.Search(context.Background(), &dest, query,
			opensearch.Search.WithIndex("auditlog"),
			opensearch.Search.WithRequestCache(false),
		)
		if err != nil {
			t.Fatalf("unable to search for document")
		}
		g.Expect(len(dest)).To(gomega.Equal(5))
	}
}
