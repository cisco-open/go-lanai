package instrument

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	goredis "github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// redisTracingHook implements redis.Hook and redis.OptionsAwareHook
type redisTracingHook struct {
	tracer opentracing.Tracer
	db     int
}

func NewRedisTrackingHook(tracer opentracing.Tracer) *redisTracingHook{
	return newRedisTrackingHook(tracer, -1)
}

func newRedisTrackingHook(tracer opentracing.Tracer, db int) *redisTracingHook{
	return &redisTracingHook{
		tracer: tracer,
		db: -1,
	}
}

// redis.OptionsAwareHook
func (h redisTracingHook) WithClientOption(opts *goredis.UniversalOptions) goredis.Hook {
	return newRedisTrackingHook(h.tracer, opts.DB)
}

// redis.Hook
func (h redisTracingHook) BeforeProcess(ctx context.Context, cmd goredis.Cmder) (context.Context, error) {
	name := tracing.OpNameRedis + " " + cmd.Name()
	cmdStr := cmd.Name()
	opts := []tracing.SpanOption{
		tracing.SpanKind(ext.SpanKindRPCClientEnum),
		tracing.SpanTag("cmd", cmdStr),
	}
	if h.db >= 0 {
		opts = append(opts, tracing.SpanTag("db", h.db))
	}

	return tracing.WithTracer(h.tracer).
		WithOpName(name).
		WithOptions(opts...).
		DescendantOrNoSpan(ctx), nil
}

// redis.Hook
func (h redisTracingHook) AfterProcess(ctx context.Context, cmd goredis.Cmder) error {
	op := tracing.WithTracer(h.tracer)
	if cmd.Err() != nil {
		op.WithOptions(tracing.SpanTag("err", cmd.Err()))
	}
	op.Finish(ctx)
	return nil
}

// redis.Hook
func (h redisTracingHook) BeforeProcessPipeline(ctx context.Context, cmds []goredis.Cmder) (context.Context, error) {
	name := tracing.OpNameRedis + "-batch"
	cmdNames := make([]string, len(cmds))
	for i, v := range cmds {
		cmdNames[i] = v.Name()
	}
	opts := []tracing.SpanOption{
		tracing.SpanKind(ext.SpanKindRPCClientEnum),
		tracing.SpanTag("cmd", cmdNames),
	}
	if h.db >= 0 {
		opts = append(opts, tracing.SpanTag("data", h.db))
	}
	return tracing.WithTracer(h.tracer).
		WithOpName(name).
		WithOptions(opts...).
		DescendantOrNoSpan(ctx), nil
}

// redis.Hook
func (h redisTracingHook) AfterProcessPipeline(ctx context.Context, cmds []goredis.Cmder) error {
	op := tracing.WithTracer(h.tracer)
	errs := map[string]error{}
	for _, v := range cmds {
		if v.Err() != nil {
			errs[v.Name()] = v.Err()
		}
	}
	if len(errs) != 0 {
		op.WithOptions(tracing.SpanTag("err", errs))
	}
	op.Finish(ctx)
	return nil
}
