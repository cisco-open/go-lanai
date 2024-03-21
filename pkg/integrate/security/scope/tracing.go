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

package scope

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/fx"
)

const tracingOpName = "security"

type tracingManagerCustomizer struct {
	tracer opentracing.Tracer
}

func tracingProvider() fx.Annotated {
	return FxManagerCustomizer(newSecurityScopeManagerCustomizer)
}

type tracingDI struct {
	fx.In
	Tracer opentracing.Tracer `optional:"true"`
}

func newSecurityScopeManagerCustomizer(di tracingDI) ManagerCustomizer {
	return &tracingManagerCustomizer{
		tracer: di.Tracer,
	}
}

func (c *tracingManagerCustomizer) Customize() []ManagerOptions {
	if c.tracer == nil {
		return []ManagerOptions{}
	}
	return []ManagerOptions{
		BeforeStartHook(startSpanHook(c.tracer)),
		AfterEndHook(finishSpanHook(c.tracer)),
	}
}

func startSpanHook(tracer opentracing.Tracer) ScopeOperationHook {
	return func(ctx context.Context, scope *Scope) context.Context {
		name := tracingOpName
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

func finishSpanHook(tracer opentracing.Tracer) ScopeOperationHook {
	return func(ctx context.Context, _ *Scope) context.Context {
		return tracing.WithTracer(tracer).FinishAndRewind(ctx)
	}
}

