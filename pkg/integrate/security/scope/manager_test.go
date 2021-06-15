package scope

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	webinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	test "cto-github.cisco.com/NFV-BU/go-lanai/test/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/utils/testapp"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/utils/testsuite"
	"fmt"
	"github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
	"time"
)

type testHookCounter struct {
	setupCount       int
	subSetupCount    int
	teardownCount    int
	subTeardownCount int
	fxInvokeCount    int
}

func (c *testHookCounter) setup(ctx context.Context, _ *testing.T) (context.Context, error) {
	c.setupCount++
	return ctx, nil
}

func (c *testHookCounter) teardown(_ *testing.T) error {
	c.teardownCount++
	return nil
}

func (c *testHookCounter) subSetup(ctx context.Context, _ *testing.T) (context.Context, error) {
	c.subSetupCount++
	return ctx, nil
}

func (c *testHookCounter) subTeardown(_ *testing.T) error {
	c.subTeardownCount++
	return nil
}

func (c *testHookCounter) fxInvoke(_ webDI) error {
	c.fxInvokeCount++
	return nil
}

type bootstrapDI struct {
	fx.In
	App *bootstrap.App
}

type appconfigDI struct {
	fx.In
	ACPtr *appconfig.ApplicationConfig
	ACI   bootstrap.ApplicationConfig
}

type webDI struct {
	fx.In
	Register *web.Registrar
}

type controller struct{}

func newController() web.Controller {
	return &controller{}
}

func (c *controller) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.New("test").Get("/api").EndpointFunc(c.Test).Build(),
	}
}

func (c *controller) Test(_ context.Context, _ *http.Request) (response interface{}, err error) {
	return map[string]string{
		"message": "ok",
	}, nil
}

/*************************
	Test Main Setup
 *************************/

func TestMain(m *testing.M) {
	testsuite.RunTests(m,
		testsuite.TestOptions(testapp.WithModules(webinit.Module)),
	)
}

/*************************
	Test Cases
 *************************/
func TestBootstrapWithDefaults(t *testing.T) {
	counter := &testHookCounter{}
	bDI := &bootstrapDI{}
	acDI := &appconfigDI{}
	test.RunTest(context.Background(), t,
		testapp.Bootstrap(),
		testapp.WithDI(bDI, acDI),
		test.Setup(counter.setup),
		test.Teardown(counter.teardown),
		test.SubTestSetup(counter.subSetup),
		test.SubTestTeardown(counter.subTeardown),
		test.GomegaSubTest(SubTestDefaultDI(bDI, acDI), "SubTestDefaultDI-Pass1"),
		test.GomegaSubTest(SubTestDefaultDI(bDI, acDI), "SubTestDefaultDI-Pass2"),
		test.GomegaSubTest(SubTestDefaultDI(bDI, acDI), "SubTestDefaultDI-Pass3"),
	)

	g := gomega.NewWithT(t)
	g.Expect(counter.setupCount).To(gomega.Equal(1), "Test setup should invoked once per test")
	g.Expect(counter.teardownCount).To(gomega.Equal(1), "Test teardown should invoked once per test")
	g.Expect(counter.subSetupCount).To(gomega.Equal(3), "SubTest setup should invoked once per sub test")
	g.Expect(counter.subTeardownCount).To(gomega.Equal(3), "SubTest teardown should invoked once per sub test")
	g.Expect(counter.fxInvokeCount).To(gomega.Equal(0), "fx invoke func should be be triggerred")
}

func TestBootstrapWithCustomSettings(t *testing.T) {
	counter := &testHookCounter{}
	bDI := &bootstrapDI{}
	acDI := &appconfigDI{}
	wDI := &webDI{}
	test.RunTest(context.Background(), t,
		testapp.Bootstrap(),
		testapp.WithTimeout(30*time.Second),
		testapp.WithFxPriorityOptions(
			fx.Invoke(counter.fxInvoke),
		),
		testapp.WithFxOptions(
			web.FxControllerProviders(newController),
			fx.Invoke(counter.fxInvoke),
		),
		testapp.WithDI(bDI, acDI, wDI),
		test.GomegaSubTest(SubTestDefaultDI(bDI, acDI)),
		test.GomegaSubTest(SubTestAdditionalDI(wDI)),
		test.GomegaSubTest(SubTestWebController(wDI)),
	)

	g := gomega.NewWithT(t)
	g.Expect(counter.fxInvokeCount).To(gomega.Equal(2), "fx invoke func should be invoked twice, 1 for regular order, 1 for priority order")
}

/*************************
	Sub-Test Cases
 *************************/
func SubTestDefaultDI(bDI *bootstrapDI, acDI *appconfigDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(bDI.App).To(gomega.Not(gomega.BeNil()), "bootstrap DI should be populated with App")
		g.Expect(acDI.ACPtr).To(gomega.Not(gomega.BeNil()), "appconfig DI should be populated with *appconfig.ApplicationConfig")
		g.Expect(acDI.ACI).To(gomega.Not(gomega.BeNil()), "appconfig DI should be populated with bootstrap.ApplicationConfig")
	}
}

func SubTestAdditionalDI(di *webDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Register).To(gomega.Not(gomega.BeNil()), "web DI should be populated with Registrar")
	}
}

func SubTestWebController(di *webDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		port := di.Register.ServerPort()
		url := fmt.Sprintf("http://localhost:%d/test/api", port)
		resp, e := http.DefaultClient.Get(url)
		g.Expect(e).To(gomega.Succeed(), "http client should be succeeded")
		g.Expect(resp).To(gomega.Not(gomega.BeNil()), "http response should not be nil ")
		g.Expect(resp.StatusCode).To(gomega.Equal(200), "http response status should be 200")
	}
}

/*************************
	Helpers
 *************************/

//func (c *cache) loadFunc() loadFunc {
//	return func(ctx context.Context, k cKey) (v entryValue, exp time.Time, err error) {
//		fmt.Printf("loading key-%v...\n", k)
//		time.Sleep(1 * time.Second)
//		if k % 2 == 0 {
//			// happy path valid 5 seconds
//			valid := 5 * time.Second
//			exp = time.Now().Add(valid)
//			v = oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
//				opt.Token = oauth2.NewDefaultAccessToken("My test token")
//			})
//			fmt.Printf("loaded key-%v=%v with exp in %v\n", k, v.AccessToken().Value(), valid)
//		} else {
//			// unhappy path valid 2 seconds
//			valid := 2 * time.Second
//			exp = time.Now().Add(valid)
//			err = fmt.Errorf("oops")
//			fmt.Printf("loaded key-%v=%v with exp in %v\n", k, err, valid)
//		}
//		return
//	}
//}

//func (c *cache) evictFunc() gcache.EvictedFunc {
//	return func(k interface{}, v interface{}) {
//		fmt.Printf("evicted %v=%v\n", k, v)
//	}
//}
