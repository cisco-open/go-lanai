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

