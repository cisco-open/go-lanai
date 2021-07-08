package examples

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	webinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"embed"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
	"time"
)

/*************************
	Setup
 *************************/

type DummyService interface {
	DummyMethod(_ context.Context) error
}

func NewRealService() DummyService {
	return &realService{}
}

type realService struct {}

func (t *realService) DummyMethod(_ context.Context) error {
	return nil
}

type DummyController struct{
	svc DummyService
}

func NewDummyController(svc DummyService) web.Controller {
	return &DummyController{
		svc: svc,
	}
}

func (c *DummyController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.New("test").Get("/api").EndpointFunc(c.Test).Build(),
	}
}

func (c *DummyController) Test(_ context.Context, _ *http.Request) (response interface{}, err error) {
	return map[string]string{
		"message": "ok",
	}, nil
}

type serviceDI struct {
	fx.In
	Service DummyService
}

type webDI struct {
	fx.In
	Register *web.Registrar
}

/*************************
	Examples
 *************************/

// TestBootstrapWithDefaults
// Simple configuration to setup a set of sub tests with FX and bootstrapping
// Any number of DI struct pointers (with fx.In) can be specified.
// All specified DI struct via apptest.WithDI will be populated when sub tests is run
// Only bootstrap and appconfig package is initialized by default
func TestBootstrapWithDefaults(t *testing.T) {
	di := serviceDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(&di), 	// tell test framework to do dependencies injection
		apptest.WithFxOptions(
			fx.Provide(NewRealService), // provide real service
		),
		test.GomegaSubTest(SubTestExampleWithRealService(&di), "SubTestWithRealService"),
	)
}

//go:embed example-test-config.yml
var customConfigFS embed.FS

// TestBootstrapWithCustomConfigAndMocks
// Mocked services can be provided via apptest.WithFxOptions.
// But due to limitation of uber.fx, no provider overwriting is supported
// Config FS is loaded as application ad-hoc configs, which has higher priority than all but CLI/consul/vault config
func TestBootstrapWithCustomConfigAndMocks(t *testing.T) {
	di := serviceDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(&di), 	// tell test framework to do dependencies injection
		apptest.WithTimeout(30*time.Second),
		apptest.WithConfigFS(customConfigFS),
		apptest.WithProperties("info.inline: value", "info.inline-alt=value"),
		apptest.WithFxOptions(
			fx.Provide(NewMockedService), // provide real service
		),
		test.GomegaSubTest(SubTestExampleWithMockedService(&di), "SubTestWithMockedService"),
	)
}

// TestBootstrapWithRealWebServer
// Without specifying server port when webinit.Module is enabled, the web package would create a real
// server with random port. The port can be retrieved via *web.Registrar
// Note: due to current bootstrapping limitation, all modules added via apptest.WithModules would affect other tests,
//		 so no conflicting modules is allowed between all tests of same package.
func TestBootstrapWithRealWebServer(t *testing.T) {
	di := webDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(&di), 	// tell test framework to do dependencies injection
		apptest.WithModules(webinit.Module),
		apptest.WithFxOptions(
			// provide DummyController and mocked service
			fx.Provide(NewMockedService),
			web.FxControllerProviders(NewDummyController),
		),
		test.GomegaSubTest(SubTestExampleWithRealWebController(&di), "SubTestWithRealWebController"),
	)
}

// TestBootstrapWithRealWebServer
// Without specifying server port when webinit.Module is enabled, the web package would create a real
// server with random port. The port can be retrieved via *web.Registrar
// Note: due to current bootstrapping limitation, all modules added via apptest.WithModules would affect other tests,
//		 so no conflicting modules is allowed between all tests of same package.
func TestBootstrapWithDataMocks(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithNoopMocks(),
		test.GomegaSubTest(SubTestExampleWithOverriddenTxManager(), "SubTestWithOverriddenTxManager"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestExampleWithRealService(di *serviceDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Service).To(Not(BeNil()), "Service should be injected")
		g.Expect(di.Service).To(BeAssignableToTypeOf(&realService{}), "Injected service should be the real service")
	}
}

func SubTestExampleWithMockedService(di *serviceDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Service).To(Not(BeNil()), "Service should be injected")
		g.Expect(di.Service).To(BeAssignableToTypeOf(&mockedService{}), "Injected service should be the real service")
	}
}

func SubTestExampleWithRealWebController(di *webDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		port := di.Register.ServerPort()
		url := fmt.Sprintf("http://localhost:%d/test/api", port)
		resp, e := http.DefaultClient.Get(url)
		g.Expect(e).To(Succeed(), "http client should be succeeded")
		g.Expect(resp).To(Not(BeNil()), "http response should not be nil ")
		g.Expect(resp.StatusCode).To(Equal(200), "http response status should be 200")
	}
}

func SubTestExampleWithOverriddenTxManager() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// Regular usage
		e := tx.Transaction(ctx, func(txCtx context.Context) error {
			_, ok := txCtx.(tx.TxContext)
			g.Expect(ok).To(BeTrue(), "Overridden TxManager should create tx.TxContext")
			_, ok = txCtx.(tx.GormContext)
			g.Expect(ok).To(BeTrue(), "Overridden TxManager should create tx.GormContext")
			return nil
		})
		g.Expect(e).To(Succeed(), "Overridden TxManager shouldn't return error")

		// Manual usage
		txCtx, e := tx.Begin(ctx)
		g.Expect(e).To(Succeed(), "Overridden ManualTxManager shouldn't return error")
		_, ok := txCtx.(tx.TxContext)
		g.Expect(ok).To(BeTrue(), "Overridden TxManager should create tx.TxContext")
		_, ok = txCtx.(tx.GormContext)
		g.Expect(ok).To(BeTrue(), "Overridden TxManager should create tx.GormContext")

		txCtx, e = tx.Commit(txCtx)
		g.Expect(e).To(Succeed(), "Overridden ManualTxManager shouldn't return error")
		g.Expect(txCtx).To(BeIdenticalTo(ctx), "Overridden ManualTxManager shouldn't do anything")
	}
}