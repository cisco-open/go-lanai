package tracing_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	tracinginit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Setup Test
 *************************/

func SetupBootstrapTracing() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		tracinginit.EnableBootstrapTracing(apptest.TestBootstrapper(ctx))
		return ctx, nil
	}
}

/*************************
	Tests
 *************************/

type TestTracerDI struct {
	fx.In
	AppContext *bootstrap.ApplicationContext
	Tracer     opentracing.Tracer
}

func TestTracerWithLowestRateSampler(t *testing.T) {
	di := TestTracerDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(120*time.Second),
		apptest.WithModules(tracinginit.Module),
		apptest.WithProperties(
			"tracing.sampler.lowest-per-second: 1",
			"tracing.sampler.probability: 0.5",
		),
		apptest.WithDI(&di),
		test.Setup(SetupBootstrapTracing()),
		test.GomegaSubTest(SubTestApplicationSpan(&di), "TestApplicationSpan"),
	)
}

func TestTracerWithProbabilitySampler(t *testing.T) {
	di := TestTracerDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(120*time.Second),
		apptest.WithModules(tracinginit.Module),
		apptest.WithProperties(
			"tracing.sampler.lowest-per-second: 0",
			"tracing.sampler.probability: 0.5",
		),
		apptest.WithDI(&di),
		test.Setup(SetupBootstrapTracing()),
		test.GomegaSubTest(SubTestApplicationSpan(&di), "TestApplicationSpan"),
	)
}

func TestTracerWithRateLimitSampler(t *testing.T) {
	di := TestTracerDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(120*time.Second),
		apptest.WithModules(tracinginit.Module),
		apptest.WithProperties(
			"tracing.sampler.limit-per-second: 50",
		),
		apptest.WithDI(&di),
		test.Setup(SetupBootstrapTracing()),
		test.GomegaSubTest(SubTestApplicationSpan(&di), "TestApplicationSpan"),
	)
}

func TestTracerWithNoSampler(t *testing.T) {
	di := TestTracerDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(120*time.Second),
		apptest.WithModules(tracinginit.Module),
		apptest.WithProperties(
			"tracing.sampler.enabled: false",
		),
		apptest.WithDI(&di),
		test.Setup(SetupBootstrapTracing()),
		test.GomegaSubTest(SubTestApplicationSpan(&di), "TestApplicationSpan"),
	)
}

func TestTracerWithInvalidSampler(t *testing.T) {
	di := TestTracerDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(120*time.Second),
		apptest.WithModules(tracinginit.Module),
		apptest.WithProperties(
			"tracing.sampler.limit-per-second: 0",
		),
		apptest.WithDI(&di),
		test.Setup(SetupBootstrapTracing()),
		test.GomegaSubTest(SubTestApplicationSpan(&di), "TestApplicationSpan"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestApplicationSpan(di *TestTracerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var span opentracing.Span
		span = tracing.SpanFromContext(di.AppContext)
		g.Expect(span).ToNot(BeNil(), "application span should not be nil")
		traceID := tracing.TraceIdFromContext(di.AppContext)
		g.Expect(traceID).ToNot(BeNil(), "application traceID should not be nil")
	}
}

/*************************
	Helper
 *************************/


