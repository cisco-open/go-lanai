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

package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

/**********************
	SpanOptions
 **********************/

func SpanTag(key string, v interface{}) SpanOption {
	return func(span opentracing.Span) {
		span.SetTag(key, v)
	}
}

func SpanBaggageItem(restrictedKey string, s string) SpanOption {
	return func(span opentracing.Span) {
		span.SetBaggageItem(restrictedKey, s)
	}
}

func SpanKind(v ext.SpanKindEnum) SpanOption {
	return func(span opentracing.Span) {
		ext.SpanKind.Set(span, v)
	}
}

func SpanComponent(v string) SpanOption {
	return func(span opentracing.Span) {
		ext.Component.Set(span, v)
	}
}

func SpanHttpUrl(v string) SpanOption {
	return func(span opentracing.Span) {
		ext.HTTPUrl.Set(span, v)
	}
}

func SpanHttpMethod(v string) SpanOption {
	return func(span opentracing.Span) {
		ext.HTTPMethod.Set(span, v)
	}
}

func SpanHttpStatusCode(v int) SpanOption {
	return func(span opentracing.Span) {
		ext.HTTPStatusCode.Set(span, uint16(v))
	}
}
