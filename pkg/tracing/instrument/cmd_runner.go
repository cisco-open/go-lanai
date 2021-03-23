package instrument

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/fx"
)

type cliRunnerTracingHooks struct {
	tracer opentracing.Tracer
}


func CliRunnerTracingProvider() fx.Annotated {
	return fx.Annotated{
		Group:  bootstrap.FxCliRunnerGroup,
		Target: newCliRunnerTracingHooks,
	}
}

func newCliRunnerTracingHooks(tracer opentracing.Tracer) bootstrap.CliRunnerLifecycleHooks {
	return &cliRunnerTracingHooks{tracer: tracer}
}

func (h cliRunnerTracingHooks) Before(ctx context.Context, runner bootstrap.CliRunner) context.Context {
	return tracing.WithTracer(h.tracer).
		WithOpName(tracing.OpNameCli).
		WithOptions(tracing.SpanKind(ext.SpanKindRPCServerEnum)).
		//WithOptions(tracing.SpanTag("runner", fmt.Sprintf("%v", reflect.ValueOf(runner).String()))).
		ForceNewSpan(ctx)
}

func (h cliRunnerTracingHooks) After(ctx context.Context, runner bootstrap.CliRunner, err error) context.Context {
	op := tracing.WithTracer(h.tracer)
	if err != nil {
		op = op.WithOptions(tracing.SpanTag("err", err))
	}
	return op.FinishAndRewind(ctx)
}

