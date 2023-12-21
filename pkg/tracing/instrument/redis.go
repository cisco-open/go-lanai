// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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
		db: db,
	}
}

// WithClientOption implements redis.OptionsAwareHook
func (h redisTracingHook) WithClientOption(opts *goredis.UniversalOptions) goredis.Hook {
	return newRedisTrackingHook(h.tracer, opts.DB)
}

// BeforeProcess implements redis.Hook
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

// AfterProcess implements redis.Hook
func (h redisTracingHook) AfterProcess(ctx context.Context, cmd goredis.Cmder) error {
	op := tracing.WithTracer(h.tracer)
	if cmd.Err() != nil {
		op.WithOptions(tracing.SpanTag("err", cmd.Err()))
	}
	op.Finish(ctx)
	return nil
}

// AfterProcessPipeline implements redis.Hook
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

// AfterProcessPipeline implements redis.Hook
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
