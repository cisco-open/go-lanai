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
	"github.com/cisco-open/go-lanai/pkg/integrate/security/scope"
	"github.com/cisco-open/go-lanai/pkg/tracing"
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

