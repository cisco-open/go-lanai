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
	"github.com/cisco-open/go-lanai/pkg/tracing"
	util_matcher "github.com/cisco-open/go-lanai/pkg/utils/matcher"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/opentracing/opentracing-go"
	"net/http"
	"strings"
)

var (
	excludeRequest = util_matcher.Or(&healthMatcher, &corsPreflightMatcher)
)

type TracingWebCustomizer struct {
	tracer opentracing.Tracer
}

func NewTracingWebCustomizer(tracer opentracing.Tracer) *TracingWebCustomizer{
	return &TracingWebCustomizer{
		tracer: tracer,
	}
}

// Order we want TracingWebCustomizer before anything else
func (c TracingWebCustomizer) Order() int {
	return order.Highest
}

func (c *TracingWebCustomizer) Customize(_ context.Context, r *web.Registrar) error {
	// for gin
	//nolint:contextcheck
	if e := r.AddGlobalMiddlewares(GinTracing(c.tracer, tracing.OpNameHttp, excludeRequest)); e != nil {
		return e
	}

	// for go-kit endpoints, because we are unable to finish the created span,
	// so we rely on Gin middleware to create/finish span
	//t := kithttp.ServerBefore(kitopentracing.HTTPToContext(c.tracer, tracing.OpNameHttp, logger))
	//r.AddOption(t)
	return nil
}


/*********************
	common funcs
 *********************/
func opNameWithRequest(opName string, r *http.Request) string {
	return opName + " " + r.URL.Path
}

/*********************
	exlusion matcher
 *********************/
var (
	healthMatcher = exclusionMatcher{
		matches: func(r *http.Request) bool {
			return strings.HasSuffix(r.URL.Path, "/health") && r.Method == http.MethodGet
		},
	}

	corsPreflightMatcher = exclusionMatcher{
		matches: func(r *http.Request) bool {
			return r.Method == http.MethodOptions
		},
	}
)

// exclusionMatcher is specialized web.RequestMatcher that do faster matching (simplier and relaxed logic)
type exclusionMatcher struct {
	matches func(*http.Request) bool
}

func (m exclusionMatcher) Matches(i interface{}) (bool, error) {
	r, ok := i.(*http.Request)
	return ok && m.matches(r) , nil
}

func (m exclusionMatcher) MatchesWithContext(_ context.Context, i interface{}) (bool, error) {
	return m.Matches(i)
}

func (m exclusionMatcher) Or(matcher ...util_matcher.Matcher) util_matcher.ChainableMatcher {
	return util_matcher.Or(m, matcher...)
}

func (m exclusionMatcher) And(matcher ...util_matcher.Matcher) util_matcher.ChainableMatcher {
	return util_matcher.And(m, matcher...)
}


