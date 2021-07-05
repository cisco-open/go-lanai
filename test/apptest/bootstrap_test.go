package apptest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	webinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
	"embed"
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

type dummyService struct{}

func newDummyService() *dummyService {
	return &dummyService{}
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
	suitetest.RunTests(m,
		suitetest.TestOptions(WithModules(webinit.Module)),
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
		Bootstrap(),
		WithDI(bDI, acDI),
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

//go:embed bootstrap_test.yml
var testConfigFS embed.FS

func TestBootstrapWithCustomSettings(t *testing.T) {
	counter := &testHookCounter{}
	bDI := &bootstrapDI{}
	acDI := &appconfigDI{}
	wDI := &webDI{}
	test.RunTest(context.Background(), t,
		Bootstrap(),
		WithTimeout(30*time.Second),
		WithConfigFS(testConfigFS),
		WithConfigFS(TestApplicationConfigFS),
		WithFxPriorityOptions(
			fx.Invoke(counter.fxInvoke),
		),
		WithFxOptions(
			web.FxControllerProviders(newController),
			fx.Provide(newDummyService),
			fx.Invoke(counter.fxInvoke),
		),
		WithDI(bDI, acDI, wDI),
		test.GomegaSubTest(SubTestDefaultDI(bDI, acDI)),
		test.GomegaSubTest(SubTestAdditionalDI(wDI)),
		test.GomegaSubTest(SubTestCustomConfig(acDI)),
		test.GomegaSubTest(SubTestWebController(wDI)),
	)

	g := gomega.NewWithT(t)
	g.Expect(counter.fxInvokeCount).To(gomega.Equal(2), "fx invoke func should be invoked twice, 1 for regular order, 1 for priority order")
}

func TestRepeatedBootstrapWithCustomSettings(t *testing.T) {
	counter := &testHookCounter{}
	bDI := &bootstrapDI{}
	acDI := &appconfigDI{}
	wDI := &webDI{}
	test.RunTest(context.Background(), t,
		Bootstrap(),
		WithTimeout(30*time.Second),
		WithProperties(
			"info.source: 200",
			"info.placeholder=${info.source}",
		),
		WithDynamicProperties(map[string]PropertyValuerFunc{
			"info.source": func(_ context.Context) interface{} {return 200.0},
		}),
		WithFxPriorityOptions(
			fx.Invoke(counter.fxInvoke),
		),
		WithFxOptions(
			web.FxControllerProviders(newController),
			fx.Provide(newDummyService),
			fx.Invoke(counter.fxInvoke),
		),
		WithDI(bDI, acDI, wDI),
		test.GomegaSubTest(SubTestDefaultDI(bDI, acDI)),
		test.GomegaSubTest(SubTestAdditionalDI(wDI)),
		test.GomegaSubTest(SubTestCustomConfig(acDI)),
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

func SubTestCustomConfig(acDI *appconfigDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		v := acDI.ACPtr.Value("info.placeholder")
		g.Expect(v).To(gomega.Equal(200.0), "web DI should be populated with Registrar")
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
