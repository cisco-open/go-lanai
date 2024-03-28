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

package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
)

/*************************
	ErrorHandlerMapping
 *************************/

type errorTranslateMapping struct {
	name          string
	order         int
	matcher       RouteMatcher
	condition     RequestMatcher
	translateFunc ErrorTranslateFunc
}

func NewErrorTranslateMapping(name string, order int, matcher RouteMatcher, cond RequestMatcher, translateFunc ErrorTranslateFunc) ErrorTranslateMapping {
	return &errorTranslateMapping{
		name:          name,
		matcher:       matcher,
		order:         order,
		condition:     cond,
		translateFunc: translateFunc,
	}
}

func (m errorTranslateMapping) Name() string {
	return m.name
}

func (m errorTranslateMapping) Matcher() RouteMatcher {
	return m.matcher
}

func (m errorTranslateMapping) Order() int {
	return m.order
}

func (m errorTranslateMapping) Condition() RequestMatcher {
	return m.condition
}

func (m errorTranslateMapping) TranslateFunc() ErrorTranslateFunc {
	return m.translateFunc
}

/*************************
	Error Translation
 *************************/

func newErrorEncoder(encoder EncodeErrorFunc, translators ...ErrorTranslator) EncodeErrorFunc {
	return func(ctx context.Context, err error, rw http.ResponseWriter) {
		for _, t := range translators {
			err = t.Translate(ctx, err)
		}
		encoder(ctx, err, rw)
	}
}

type mappedErrorTranslator struct {
	order         int
	condition     RequestMatcher
	translateFunc ErrorTranslateFunc
}

func (t mappedErrorTranslator) Order() int {
	return t.order
}

func (t mappedErrorTranslator) Translate(ctx context.Context, err error) error {
	if t.condition != nil {
		if ginCtx := GinContext(ctx); ginCtx != nil {
			if matched, e := t.condition.MatchesWithContext(ctx, ginCtx.Request); e != nil || !matched {
				return err
			}
		}
	}
	return t.translateFunc(ctx, err)
}

func newMappedErrorTranslator(m ErrorTranslateMapping) *mappedErrorTranslator {
	return &mappedErrorTranslator{
		order:         m.Order(),
		condition:     m.Condition(),
		translateFunc: m.TranslateFunc(),
	}
}

type defaultErrorTranslator struct{}

func (i defaultErrorTranslator) Translate(_ context.Context, err error) error {
	//nolint:errorlint
	switch e := err.(type) {
	case validator.ValidationErrors:
		return ValidationErrors{e}
	case StatusCoder, HttpError:
		return err
	default:
		return HttpError{error: err, SC: http.StatusInternalServerError}
	}
}

func newDefaultErrorTranslator() defaultErrorTranslator {
	return defaultErrorTranslator{}
}

/*****************************
	Error Encoder
******************************/

func JsonErrorEncoder() EncodeErrorFunc {
	return jsonErrorEncoder
}

//nolint:errorlint
func jsonErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	// body
	if _, ok := err.(json.Marshaler); !ok {
		err = NewHttpError(0, err)
	}
	body, e := json.Marshal(err)
	if e != nil {
		body = []byte(fmt.Sprintf(`{"error":"%s"}`, err.Error()))
	}
	// headers
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if headerer, ok := err.(Headerer); ok {
		for k, values := range headerer.Headers() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}
	// status code
	code := http.StatusInternalServerError
	if sc, ok := err.(StatusCoder); ok {
		code = sc.StatusCode()
	}
	// write response
	w.WriteHeader(code)
	_, _ = w.Write(body)
}
