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
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/gin-gonic/gin"
	kitopentracing "github.com/go-kit/kit/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
)

func GinTracing(tracer opentracing.Tracer, opName string, excludes web.RequestMatcher) gin.HandlerFunc {
	return func(gc *gin.Context) {
		if m, e := excludes.Matches(gc.Request); e == nil && m {
			return
		}

		orig := gc.Request.Context()

		// start or join span
		reqFunc := kitopentracing.HTTPToContext(tracer, opNameWithRequest(opName, gc.Request), logger)
		ctx := reqFunc(orig, gc.Request)
		gc.Request = gc.Request.WithContext(ctx)

		gc.Next()

		// finish the span
		tracing.WithTracer(tracer).
			WithOptions(tracing.SpanHttpStatusCode(gc.Writer.Status())).
			Finish(ctx)
		gc.Request = gc.Request.WithContext(orig)
	}
}
