package v1

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"github.com/cisco-open/go-lanai/examples/skeleton-service/pkg/api"
	v1 "github.com/cisco-open/go-lanai/examples/skeleton-service/pkg/api/v1"
	"github.com/cisco-open/go-lanai/examples/skeleton-service/pkg/model"
	"github.com/cisco-open/go-lanai/examples/skeleton-service/pkg/repository"
	"github.com/cisco-open/go-lanai/pkg/data/repo"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/dbtest"
	"github.com/cisco-open/go-lanai/test/sectest"
	"github.com/cisco-open/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
	"testing"
)

//go:embed testdata/*.sql
var testDataFS embed.FS

// Uncomment this function to enable database record mode
//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

type testDI struct {
	fx.In
	dbtest.DI
	Controller *ExampleFriendsController `optional:"true"`
}

func TestControllerDirectly(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		// This option won't create the database automatically.
		// You should create this database with this name before running the tests in record mode.
		dbtest.WithDBPlayback("testdb"),
		apptest.WithModules(repo.Module),
		apptest.WithFxOptions(
			fx.Provide(repository.NewFriendRepository),
			fx.Provide(NewExampleFriendsController),
		),
		apptest.WithDI(di),
		test.SubTestSetup(SubTestSetupPrepareTables(&di.DI)),
		test.GomegaSubTest(SubTestPostItem(di), "TestPostItem"),
	)

}

func TestControllerWithHttpRequest(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		// This option won't create the database automatically.
		// You should create this database with this name before running the tests in record mode.
		dbtest.WithDBPlayback("testdb"),
		apptest.WithModules(repo.Module),
		webtest.WithMockedServer(),
		// This option is not required if the test only need to extract the security context from the request's context.
		// Uncomment this option if the test needs to mock the security context using other approach, for example, mocking
		// the security context dynamically based on the request.
		//sectest.WithMockedMiddleware(),
		apptest.WithFxOptions(
			fx.Provide(repository.NewFriendRepository),
			web.FxControllerProviders(NewExampleFriendsController),
		),
		apptest.WithDI(di),
		test.SubTestSetup(SubTestSetupPrepareTables(&di.DI)),
		test.GomegaSubTest(SubTestPostItemRequest(di), "TestPostItem"),
	)
}

func SubTestPostItem(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(
			func(d *sectest.SecurityDetailsMock) {
				d.Username = "testuser"
			}))

		firstName := "John"
		lastName := "Doe"
		req := v1.PostItemsRequest{
			ExampleRequest: api.ExampleRequest{
				RequestItem: api.RequestItem{
					FirstName: &firstName,
					LastName:  &lastName,
				},
			},
		}
		s, resp, err := di.Controller.PostItems(ctx, req)
		g.Expect(err).To(Not(HaveOccurred()), "post item should not return error")
		g.Expect(s).To(Equal(http.StatusCreated), "post item should return 201")

		respItem, ok := resp.(api.ResponseItem)
		g.Expect(ok).To(BeTrue(), "response should be of type ResponseItem")
		g.Expect(respItem.FirstName).To(Equal(req.FirstName), "response first name should match request")
		g.Expect(respItem.LastName).To(Equal(req.LastName), "response last name should match request")

		var records []*model.Friend
		di.DB.Find(&records, model.Friend{FirstName: firstName, LastName: lastName})
		g.Expect(len(records)).To(Equal(1), "should have 1 record in the database")
		g.Expect(records[0].CreatedBy).To(Equal("testuser"), "created by should be testuser")
	}
}

func SubTestPostItemRequest(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(
			func(d *sectest.SecurityDetailsMock) {
				d.Username = "testuser"
			}))

		firstName := "John"
		lastName := "Doe"
		item := v1.PostItemsRequest{
			ExampleRequest: api.ExampleRequest{
				RequestItem: api.RequestItem{
					FirstName: &firstName,
					LastName:  &lastName,
				},
			},
		}
		b, err := json.Marshal(item)
		g.Expect(err).NotTo(HaveOccurred())
		r := bytes.NewReader(b)
		req := webtest.NewRequest(ctx, http.MethodPost, "/api/v1/example/friends", r)
		req.Header.Set("Content-Type", "application/json")

		resp, e := webtest.Exec(ctx, req)
		g.Expect(e).NotTo(HaveOccurred())
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusCreated))
		body, _ := io.ReadAll(resp.Response.Body)
		responseItem := api.ResponseItem{}
		e = json.Unmarshal(body, &responseItem)
		g.Expect(e).To(Not(HaveOccurred()))
		g.Expect(responseItem.FirstName).To(Equal(item.FirstName))
		g.Expect(responseItem.LastName).To(Equal(item.LastName))

		var records []*model.Friend
		di.DB.Find(&records, model.Friend{FirstName: firstName, LastName: lastName})
		g.Expect(len(records)).To(Equal(1), "should have 1 record in the database")
		g.Expect(records[0].CreatedBy).To(Equal("testuser"), "created by should be testuser")
	}
}

func SubTestSetupPrepareTables(di *dbtest.DI) test.SetupFunc {
	return dbtest.PrepareData(di,
		CreateAllTables(), TruncateAllTables(),
	)
}

func CreateAllTables() dbtest.DataSetupStep {
	return dbtest.SetupUsingSQLFile(testDataFS, "testdata/friends_table.sql")
}

func TruncateAllTables() dbtest.DataSetupStep {
	return dbtest.SetupTruncateTables("friends")
}
