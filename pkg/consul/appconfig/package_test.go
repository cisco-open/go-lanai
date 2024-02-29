package consulappconfig_test

import (
	"context"
	"embed"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/consul"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/consultest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

//go:embed testdata/bootstrap-test.yml
var TestBootstrapFS embed.FS

/*************************
	Test Setup
 *************************/

type TestProperties struct {
	FromLocalFile      string `json:"from-local-file"`
	FromDefault        string `json:"from-default"`
	FromDefaultProfile string `json:"from-default-profile"`
	FromApp            string `json:"from-app"`
	FromAppProfile     string `json:"from-app-profile"`
}

/*************************
	Tests
 *************************/

type TestAppConfigDI struct {
	fx.In
	Consul    *consul.Connection
	AppConfig bootstrap.ApplicationConfig
}

func TestAppConfig(t *testing.T) {
	di := TestAppConfigDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		consultest.WithHttpPlayback(t,
			//consultest.HttpRecordingMode(),
			consultest.MoreHTTPVCROptions(),
		),
		apptest.WithBootstrapConfigFS(TestBootstrapFS),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupTestPrepareProperties(&di)),
		test.GomegaSubTest(SubTestPropertiesBinding(&di), "PropertiesBinding"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SetupTestPrepareProperties(di *TestAppConfigDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		_ = di.Consul.SetKeyValue(ctx, "testconfig/default/test.from-default", []byte("default-context"))
		_ = di.Consul.SetKeyValue(ctx, "testconfig/default/test.from-default-profile", []byte("default-context"))
		_ = di.Consul.SetKeyValue(ctx, "testconfig/default/test.from-app", []byte("default-context"))
		_ = di.Consul.SetKeyValue(ctx, "testconfig/default/test.from-app-profile", []byte("default-context"))

		_ = di.Consul.SetKeyValue(ctx, "testconfig/default,testprofile/test.from-default-profile", []byte("default-context-profile"))
		_ = di.Consul.SetKeyValue(ctx, "testconfig/default,testprofile/test.from-app", []byte("default-context-profile"))
		_ = di.Consul.SetKeyValue(ctx, "testconfig/default,testprofile/test.from-app-profile", []byte("default-context-profile"))

		_ = di.Consul.SetKeyValue(ctx, "testconfig/test-app/test.from-app", []byte("test-app"))
		_ = di.Consul.SetKeyValue(ctx, "testconfig/test-app/test.from-app-profile", []byte("test-app"))

		_ = di.Consul.SetKeyValue(ctx, "testconfig/test-app,testprofile/test.from-app-profile", []byte("test-app-profile"))
		return ctx, nil
	}
}

func SubTestPropertiesBinding(di *TestAppConfigDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		p := TestProperties{}
		e := di.AppConfig.Bind(&p, "test")
		g.Expect(e).To(Succeed(), "binding properties should not fail")

		g.Expect(p.FromLocalFile).To(Equal("application-test"), "%s should be loaded and bond", "FromLocalFile")
		g.Expect(p.FromDefault).To(Equal("default-context"), "%s should be loaded and bond", "FromDefault")
		g.Expect(p.FromDefaultProfile).To(Equal("default-context-profile"), "%s should be loaded and bond", "FromDefaultProfile")
		g.Expect(p.FromApp).To(Equal("test-app"), "%s should be loaded and bond", "FromApp")
		g.Expect(p.FromAppProfile).To(Equal("test-app-profile"), "%s should be loaded and bond", "FromAppProfile")
	}
}
