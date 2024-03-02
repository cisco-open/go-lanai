package vaultappconfig_test

import (
	"context"
	"embed"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/vault"
	vaultinit "github.com/cisco-open/go-lanai/pkg/vault/init"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/ittest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
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

func RecordedVaultProvider() fx.Annotated {
	return fx.Annotated{
		Group: "vault",
		Target: func(recorder *recorder.Recorder) vault.Options {
			return func(cfg *vault.ClientConfig) error {
				recorder.SetRealTransport(cfg.HttpClient.Transport)
				cfg.HttpClient.Transport = recorder
				return nil
			}
		},
	}
}

/*************************
	Tests
 *************************/

type TestAppConfigDI struct {
	fx.In
	Vault     *vault.Client
	AppConfig bootstrap.ApplicationConfig
}

func TestAppConfig(t *testing.T) {
	di := TestAppConfigDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		ittest.WithHttpPlayback(t,
			//ittest.HttpRecordingMode(),
		),
		apptest.WithBootstrapConfigFS(TestBootstrapFS),
		apptest.WithModules(vaultinit.Module),
		apptest.WithFxOptions(
			fx.Provide(RecordedVaultProvider()),
		),
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
		_, _ = di.Vault.Logical(ctx).Write("secret/default", map[string]interface{}{
			"test.from-default":         "default-context",
			"test.from-default-profile": "default-context",
			"test.from-app":             "default-context",
			"test.from-app-profile":     "default-context",
		})

		_, _ = di.Vault.Logical(ctx).Write("secret/default/testprofile", map[string]interface{}{
			"test.from-default-profile": "default-context-profile",
			"test.from-app":             "default-context-profile",
			"test.from-app-profile":     "default-context-profile",
		})

		_, _ = di.Vault.Logical(ctx).Write("secret/test-app", map[string]interface{}{
			"test.from-app":         "test-app",
			"test.from-app-profile": "test-app",
		})

		_, _ = di.Vault.Logical(ctx).Write("secret/test-app/testprofile", map[string]interface{}{
			"test.from-app-profile": "test-app-profile",
		})
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
