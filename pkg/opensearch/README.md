# Open Search

When using the opensearch package, developers are encouraged to use WithOpenSearchPlayback, which wraps httpvcr to test. Examples can be found in the `go-lanai/pkg/test/opensearchtest/`.


When tests are run WithOpenSearchPlayback, it will generate a yml file in a `./testdata` directory that stores the outgoing requests and corresponding responses and mock them when the test suite is not in RecordMode.

BodyModifiers can be used to change expected requests/responses - like time.

The vcr will also preappend(with `test_`) any index names that are passed as options to the `Repo[T]` interface.

Check out the example test below.

```go
package main

import (
    "testing"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
    "cto-github.cisco.com/NFV-BU/go-lanai/test"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
    "go.uber.org/fx"
    "github.com/onsi/gomega"
    "github.com/opensearch-project/opensearch-go/opensearchapi"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opensearch"
    "context"
)

// To enable record mode, uncomment the block of code below
//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

type FakeService struct {
	Repo opensearch.Repo[someModel]
}

type fakeServiceDI struct {
	fx.In
	Client opensearch.OpenClient
}

func NewFakeService(di fakeServiceDI) FakeService {
	return FakeService{
		Repo: opensearch.NewRepo(&someModel{}, di.Client),
	}
}

type opensearchDI struct {
	fx.In
	FakeService   FakeService
	Properties    *opensearch.Properties
	BodyModifiers *MatcherBodyModifiers
}

func TestScopeController(t *testing.T) {
	di := &opensearchDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithOpenSearchPlayback(
			SetRecordMode(ModeCommandline),
			SetRecordDelay(time.Millisecond*1500), // Used to ensure enough time for db changes to take place
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
		//test.SubTestSetup(SetupOpenSearchTest(di)),
		test.GomegaSubTest(SubTestTemplateAndAlias(di), "SubTestNewBulkIndexer"),
	)
}


func SubTestTemplateAndAlias(di *opensearchDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		fakeNewIndexName := "generic_events_1"
		fakeIndexAlias := "generic_event"
		fakeTemplateName := "test_template"
		indexTemplate := map[string]interface{}{
			"index_patterns": []string{"*generic_events*"}, // Pattern needs to accomodate "test_" append
			"template": map[string]interface{}{
				"settings": map[string]interface{}{
					"number_of_shards":   4,
					"number_of_replicas": 4,
				},
			},
			"version": 1,
			"_meta": map[string]interface{}{
				"description": "some description",
			},
		}
		indexMapping := map[string]interface{}{
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"SubType": map[string]interface{}{
						"type": "text",
					},
				},
			},
		}
		

		// Create a Template
		err := di.FakeService.Repo.IndicesPutIndexTemplate(ctx, fakeTemplateName, indexTemplate)
		if err != nil {
			t.Fatalf("unable to create index template")
		}

		// Create an Index with template pattern
		err = di.FakeService.Repo.IndicesCreate(ctx, fakeNewIndexName, indexMapping)
		if err != nil {
			t.Fatalf("unable to create index")
		}

		// Create an Alias for the template
		err = di.FakeService.Repo.IndicesPutAlias(ctx, []string{fakeNewIndexName}, fakeIndexAlias)
		if err != nil {
			t.Fatalf("unable to create alias ")
		}

		// Get the new index using the Alias and check the obj
		resp, err := di.FakeService.Repo.IndicesGet(ctx, fakeIndexAlias)
		if err != nil {
			t.Fatalf("unable to get indices information using alias ")
		}

		// This test proves that the index template works against the newly created indices
		g.Expect(resp.Settings.Index.NumberOfShards).To(gomega.Equal("4"))

		// Test Cleanup
		// Delete Alias
		err = di.FakeService.Repo.IndicesDeleteAlias(ctx, []string{fakeNewIndexName}, []string{fakeIndexAlias})
		if err != nil {
			t.Fatalf("unable to delete indices alias ")
		}
		// Delete Index Template
		err = di.FakeService.Repo.IndicesDeleteIndexTemplate(ctx, fakeTemplateName)
		if err != nil {
			t.Fatalf("unable to delete index template ")
		}
		// Delete index
		err = di.FakeService.Repo.IndicesDelete(ctx, []string{fakeNewIndexName})
		if err != nil {
			t.Fatalf("unable to delete index ")
		}
	}
}

```

