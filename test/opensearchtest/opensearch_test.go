package opensearchtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opensearch"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"github.com/onsi/gomega"
	"go.uber.org/fx"
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
	FakeService FakeService
	Properties  *opensearch.Properties
}

func TestScopeController(t *testing.T) {
	di := &opensearchDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithOpenSearchPlayback(ModeCommandline, time.Millisecond*1500),
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
	)
}

func SetupOpenSearchTest(di *opensearchDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		return SetupPrepareOpenSearchData(
			ctx,
			di.FakeService.Repo,
			time.Date(2020, 0, 0, 0, 0, 0, 0, time.Local),
			time.Date(2022, 0, 0, 0, 0, 0, 0, time.Local),
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
		err := di.FakeService.Repo.Search(context.Background(), &dest, query,
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
			Time:      time.Date(2019, 10, 15, 0, 0, 0, 0, time.Local),
		}

		err = di.FakeService.Repo.Index(context.Background(), "auditlog", testEvent)
		if err != nil {
			t.Fatalf("unable to create document in index: %v", err)
		}
		err = di.FakeService.Repo.Search(context.Background(), &dest, query,
			opensearch.Search.WithIndex("auditlog"),
		)
		if err != nil {
			t.Fatalf("unable to search for document: %v", err)
		}
		g.Expect(len(dest)).To(gomega.Equal(3))
		g.Expect(dest[2].Client_ID).To(gomega.Equal(testEvent.Client_ID))
	}
}
