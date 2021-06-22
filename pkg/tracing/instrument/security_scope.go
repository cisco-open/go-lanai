package instrument

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/fx"
)

type securityScopeManagerCustomizer struct {
	tracer opentracing.Tracer
}

func SecurityScopeTracingProvider() fx.Annotated {
	return scope.FxManagerCustomizers(newSecurityScopeManagerCustomizer)[0]
}

func newSecurityScopeManagerCustomizer(tracer opentracing.Tracer) scope.ManagerCustomizer {
	return &securityScopeManagerCustomizer{
		tracer: tracer,
	}
}

func (s *securityScopeManagerCustomizer) Customize() []scope.ManagerOptions {
	return []scope.ManagerOptions {
		scope.BeforeStartHook(BeforeStartHook(s.tracer)),
		scope.AfterEndHook(AfterEndHook(s.tracer)),
	}
}

func BeforeStartHook(tracer opentracing.Tracer) scope.ScopeOperationHook {
	return func(ctx context.Context, scope *scope.Scope) context.Context {
		name := tracing.OpNameSecScope
		opts := []tracing.SpanOption{
			tracing.SpanKind(ext.SpanKindRPCServerEnum),
		}
		if scope != nil {
			opts = append(opts, tracing.SpanTag("sec.scope", scope.String()))
		}

		return tracing.WithTracer(tracer).
			WithOpName(name).
			WithOptions(opts...).
			DescendantOrNoSpan(ctx)
	}
}

func AfterEndHook(tracer opentracing.Tracer) scope.ScopeOperationHook {
	return func(ctx context.Context, _ *scope.Scope) context.Context {
		return tracing.WithTracer(tracer).FinishAndRewind(ctx)
	}
}

