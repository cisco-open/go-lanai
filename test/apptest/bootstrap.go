package apptest

import (
	"context"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"embed"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"testing"
)

//go:embed test-defaults.yml
var TestDefaultConfigFS embed.FS

//go:embed test-bootstrap.yml
var TestBootstrapConfigFS embed.FS

//go:embed test-application.yml
var TestApplicationConfigFS embed.FS

// testBootstrapper holds all configuration to bootstrap a fs-enabled test
type testBootstrapper struct {
	bootstrap.Bootstrapper
	AppPriorityOptions []fx.Option
	AppOptions         []fx.Option
}

// Bootstrap is an entrypoint test.Options that indicates all sub tests should be run within the scope of
// an slim version of bootstrap.App
func Bootstrap() test.Options {
	return test.WithInternalRunner(NewFxTestRunner())
}

// NewFxTestRunner is internal use only, exported for cross-package reference
func NewFxTestRunner() test.InternalRunner {
	return func(ctx context.Context, t *test.T) {
		// run setup hooks
		ctx = testSetup(ctx, t.T, t.TestHooks)
		defer testTeardown(ctx, t.T, t.TestHooks)

		// register test module's options without register the module directly
		// Note:
		//		we want to support repeated bootstrap but the bootstrap package doesn't support
		// 		module refresh (caused by singleton pattern).
		// Note 4.3:
		//		Now with help of bootstrap.ExecuteContainedApp(), we are able repeatedly bootstrap a self-contained
		// 		application.
		tb, ok := ctx.Value(ctxKeyTestBootstrapper).(*testBootstrapper)
		if !ok || tb == nil {
			t.Fatalf("Failed to start test %s due to invalid test bootstrap configuration", t.Name())
			return
		}

		// default modules
		tb.Register(appconfig.ConfigModule)

		// prepare bootstrap fx options
		priority := append([]fx.Option{
			fx.Supply(t),
			appconfig.FxEmbeddedDefaults(TestDefaultConfigFS),
			appconfig.FxEmbeddedBootstrapAdHoc(TestBootstrapConfigFS),
			appconfig.FxEmbeddedApplicationAdHoc(TestApplicationConfigFS),
		}, tb.AppPriorityOptions...)
		regular := append([]fx.Option{}, tb.AppOptions...)

		// bootstrapping
		bootstrap.NewAppCmd("testapp", priority, regular,
			func(cmd *cobra.Command) {
				cmd.Use = "testapp"
				cmd.Args = nil
			},
		)
		tb.EnableCliRunnerMode(newTestCliRunner)
		bootstrap.ExecuteContainedApp(ctx, &tb.Bootstrapper)
	}
}

func newTestCliRunner(t *test.T) bootstrap.CliRunner {
	return func(ctx context.Context) error {
		// run test
		test.InternalRunSubTests(ctx, t)
		// Note: in case of failed tests, we don't return error. GO's testing framework should be able to figure it out from t.Failed()
		return nil
	}
}

func testSetup(ctx context.Context, t *testing.T, hooks []test.Hook) context.Context {
	// run setup hooks
	for _, h := range hooks {
		var e error
		ctx, e = h.Setup(ctx, t)
		if e != nil {
			t.Fatalf("error when setup test: %v", e)
		}
	}
	return ctx
}

func testTeardown(ctx context.Context, t *testing.T, hooks []test.Hook) {
	// register cleanup
	for i := len(hooks) - 1; i >= 0; i-- {
		if e := hooks[i].Teardown(ctx, t); e != nil {
			t.Fatalf("error when setup test: %v", e)
		}
	}
}
