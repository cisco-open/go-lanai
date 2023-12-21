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
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type vaultTracingHook struct {
	tracer opentracing.Tracer
}

func NewVaultTracingHook(tracer opentracing.Tracer) *vaultTracingHook {
	return &vaultTracingHook{
		tracer: tracer,
	}
}

func (v *vaultTracingHook) BeforeOperation(ctx context.Context, cmd string) context.Context {
	name := tracing.OpNameVault + " " + cmd
	opts := []tracing.SpanOption{
		tracing.SpanKind(ext.SpanKindRPCClientEnum),
		tracing.SpanTag("cmd", cmd),
	}
	return tracing.WithTracer(v.tracer).
		WithOpName(name).
		WithOptions(opts...).
		DescendantOrNoSpan(ctx)
}

func (v *vaultTracingHook) AfterOperation(ctx context.Context, err error)  {
	op := tracing.WithTracer(v.tracer)
	if err != nil {
		op.WithOptions(tracing.SpanTag("err", err))
	}
	op.Finish(ctx)
}
