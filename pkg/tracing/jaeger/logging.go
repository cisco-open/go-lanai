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

package jaegertracing

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/uber/jaeger-client-go"
)

func init() {
	tracing.DefaultLogValuers = tracing.LogValuers{
		TraceIDValuer:  traceIdContextValuer,
		SpanIDValuer:   spanIdContextValuer,
		ParentIDValuer: parentIdContextValuer,
	}
}

func traceIdContextValuer(ctx context.Context) (ret interface{}) {
	span := tracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	switch span.Context().(type) {
	case jaeger.SpanContext:
		ret = jaegerTraceIdString(span.Context().(jaeger.SpanContext).TraceID())
	default:
		return
	}
	return
}

func spanIdContextValuer(ctx context.Context) (ret interface{}) {
	span := tracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	switch span.Context().(type) {
	case jaeger.SpanContext:
		ret = jaegerSpanIdString(span.Context().(jaeger.SpanContext).SpanID())
	}
	return
}

func parentIdContextValuer(ctx context.Context) (ret interface{}) {
	span := tracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	switch span.Context().(type) {
	case jaeger.SpanContext:
		ret = jaegerSpanIdString(span.Context().(jaeger.SpanContext).ParentID())
	default:
		return
	}
	return
}

func jaegerTraceIdString(id jaeger.TraceID) string {
	if !id.IsValid() {
		return ""
	}
	if id.High == 0 {
		return fmt.Sprintf("%.16x", id.Low)
	}
	return fmt.Sprintf("%.16x%016x", id.High, id.Low)
}

func jaegerSpanIdString(id jaeger.SpanID) string {
	if id != 0 {
		return fmt.Sprintf("%.16x", uint64(id))
	}
	return ""
}