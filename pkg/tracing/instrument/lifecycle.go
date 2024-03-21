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
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	opNameBootstrap = "bootstrap"
	opNameStart     = "startup"
	opNameStop      = "shutdown"
)

func EnableBootstrapTracing(bootstrapper *bootstrap.Bootstrapper, tracer opentracing.Tracer) {
	bootstrapper.AddInitialAppContextOptions(MakeBootstrapTracingOption(tracer, opNameBootstrap))
	bootstrapper.AddStartContextOptions(MakeStartTracingOption(tracer, opNameStart))
	bootstrapper.AddStopContextOptions(MakeStopTracingOption(tracer, opNameStop))
}

func MakeBootstrapTracingOption(tracer opentracing.Tracer, opName string) bootstrap.ContextOption {
	return func(ctx context.Context) context.Context {
		return tracing.WithTracer(tracer).
			WithOpName(opName).
			WithOptions(tracing.SpanKind(ext.SpanKindRPCServerEnum)).
			NewSpanOrDescendant(ctx)
	}
}

func MakeStartTracingOption(tracer opentracing.Tracer, opName string) bootstrap.ContextOption {
	return func(ctx context.Context) context.Context {
		return tracing.WithTracer(tracer).
			WithOpName(opName).
			WithOptions(tracing.SpanKind(ext.SpanKindRPCServerEnum)).
			NewSpanOrDescendant(ctx)
	}
}

func MakeStopTracingOption(tracer opentracing.Tracer, opName string) bootstrap.ContextOption {
	return func(ctx context.Context) context.Context {
		// finish current if not root span and start a new child
		return tracing.WithTracer(tracer).
			WithOpName(opName).
			WithOptions(tracing.SpanKind(ext.SpanKindRPCServerEnum)).
			NewSpanOrDescendant(ctx)
	}
}

