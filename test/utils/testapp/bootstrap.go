package testapp

import (
	"context"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	test "cto-github.cisco.com/NFV-BU/go-lanai/test/utils"
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

func Bootstrap() test.Options {
	return test.WithInternalRunner(NewFxTestRunner())
}

func NewFxTestRunner() test.InternalRunner {
	return func(ctx context.Context, t *test.T) {
		// run setup hooks
		ctx = testSetup(ctx, t.T, t.TestHooks)
		defer testTeardown(t.T, t.TestHooks)

		// default modules
		appconfig.Use()

		// prepare bootstrap fx options
		priority := []fx.Option{
			fx.Supply(t),
			appconfig.FxEmbeddedDefaults(TestDefaultConfigFS),
			appconfig.FxEmbeddedBootstrapAdHoc(TestBootstrapConfigFS),
			appconfig.FxEmbeddedApplicationAdHoc(TestApplicationConfigFS),
		}
		regular := make([]fx.Option, 0)

		// register test module's options without register the module directly
		// Note: we want to support repeated bootstrap but the bootstrap package doesn't support
		// module refresh (caused by singleton pattern).
		// To avoid re-declaring same fx.Option, we don't register test module directly
		if m, ok := ctx.Value(ctxKeyTestModule).(*bootstrap.Module); ok && m != nil {
			priority = append(priority, m.PriorityOptions...)
			regular = append(regular, m.Options...)
		}

		// bootstrapping
		bootstrap.NewAppCmd("testapp", priority, regular,
			func(cmd *cobra.Command) {
				cmd.Use = "testapp"
				cmd.Args = nil
			},
		)
		bootstrap.EnableCliRunnerMode(newTestCliRunner)
		bootstrap.Execute()
	}
}

func newTestCliRunner(t *test.T) bootstrap.CliRunner {
	return func(ctx context.Context) error {
		// run test
		test.InternalRunSubTests(ctx, t)
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

func testTeardown(t *testing.T, hooks []test.Hook) {
	// register cleanup
	for i := len(hooks) - 1; i >= 0; i-- {
		if e := hooks[i].Teardown(t); e != nil {
			t.Fatalf("error when setup test: %v", e)
		}
	}
}
