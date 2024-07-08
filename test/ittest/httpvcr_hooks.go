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

package ittest

import (
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/spyzhov/ajson"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultHost = "webservice"
)

const (
	HookNameIndexAware       = "index-aware"
	HookNameSanitize         = "sanitize"
	HookNameFixedDuration    = "fixed-duration"
	HookNameLocalhostRewrite = "localhost-rewrite"
)

/************************
	Common
 ************************/

func NewRecorderHook(name string, fn recorder.HookFunc, kind recorder.HookKind) RecorderHook {
	return recorderHook{
		name: name,
		hook: recorder.Hook{
			Handler: fn,
			Kind:    kind,
		},
	}
}

type recorderHook struct {
	name string
	hook recorder.Hook
}

func (h recorderHook) Name() string {
	return h.name
}

func (h recorderHook) Handler() recorder.HookFunc {
	return h.hook.Handler
}

func (h recorderHook) Kind() recorder.HookKind {
	return h.hook.Kind
}

func NewRecorderHookWithOrder(name string, fn recorder.HookFunc, kind recorder.HookKind, order int) RecorderHook {
	return orderedRecorderHook{
		recorderHook: recorderHook{
			name: name,
			hook: recorder.Hook{
				Handler: fn,
				Kind:    kind,
			},
		},
		order: order,
	}
}

type orderedRecorderHook struct {
	recorderHook
	order int
}

func (w orderedRecorderHook) Order() int {
	return w.order
}

/************************
	Sanitizer
 ************************/

var (
	HeaderSanitizers = map[string]ValueSanitizer{
		"Authorization": RegExpValueSanitizer("^(?P<prefix>Basic |Bearer |Digest ).*|.*", "${prefix}******"),
		"Date":          SubstituteValueSanitizer("Fri, 19 Aug 2022 8:51:32 GMT"),
	}
	QuerySanitizers = map[string]ValueSanitizer{
		"password":     DefaultValueSanitizer(),
		"secret":       DefaultValueSanitizer(),
		"nonce":        DefaultValueSanitizer(),
		"token":        DefaultValueSanitizer(),
		"access_token": DefaultValueSanitizer(),
	}
	BodySanitizers = map[string]ValueSanitizer{
		"access_token": DefaultValueSanitizer(),
	}
)

type ValueSanitizer func(any) any

func RegExpValueSanitizer(regex, repl string) ValueSanitizer {
	pattern := regexp.MustCompile(regex)
	return func(i any) any {
		switch s := i.(type) {
		case string:
			return pattern.ReplaceAllString(s, repl)
		default:
			return i
		}
	}
}

func SubstituteValueSanitizer(repl any) ValueSanitizer {
	return func(_ any) any {
		return repl
	}
}

func DefaultValueSanitizer() ValueSanitizer {
	return SubstituteValueSanitizer("_hidden")
}

/************************
	Hooks Functions
 ************************/

// InteractionIndexAwareHook inject interaction index into stored header:
// httpvcr store interaction's ID but doesn't expose it to cassette.MatcherFunc,
// so we need to store it in request for request matchers to access
func InteractionIndexAwareHook() RecorderHook {
	fn := func(i *cassette.Interaction) error {
		i.Request.Headers.Set(xInteractionIndexHeader, strconv.Itoa(i.ID))
		return nil
	}
	return NewRecorderHook(HookNameIndexAware, fn, recorder.BeforeSaveHook)
}

// SanitizingHook is an HTTP VCR hook that sanitize values in header, query, body (x-form-urlencoded/json).
// Values to sanitize are globally configured via HeaderSanitizers, QuerySanitizers, BodySanitizers.
// Note: Sanitized values cannot be exactly matched. If the configuration of sanitizers is changed, make sure
//
//	to configure fuzzy matching accordingly.
//
// See NewRecordMatcher, FuzzyHeaders, FuzzyQueries, FuzzyForm and FuzzyJsonPaths
func SanitizingHook() RecorderHook {
	reqJsonPaths := parseJsonPaths(FuzzyRequestJsonPaths.Values())
	respJsonPaths := parseJsonPaths(FuzzyResponseJsonPaths.Values())
	fn := func(i *cassette.Interaction) error {
		i.Request.Headers = sanitizeHeaders(i.Request.Headers, FuzzyRequestHeaders)
		i.Request.URL = sanitizeUrl(i.Request.URL, FuzzyRequestQueries)
		switch mediaType(i.Request.Headers) {
		case "application/x-www-form-urlencoded":
			i.Request.Body = sanitizeRequestForm(&i.Request, FuzzyRequestQueries)
		case "application/json":
			i.Request.Body = sanitizeJsonBody(i.Request.Body, BodySanitizers, reqJsonPaths)
		}

		i.Response.Headers = sanitizeHeaders(i.Response.Headers, FuzzyResponseHeaders)
		switch mediaType(i.Response.Headers) {
		case "application/json":
			i.Response.Body = sanitizeJsonBody(i.Response.Body, BodySanitizers, respJsonPaths)
		}
		return nil
	}
	return NewRecorderHookWithOrder(HookNameSanitize, fn, recorder.BeforeSaveHook, 0)
}

// LocalhostRewriteHook changes the host of request to a pre-defined constant if it is localhost, in order to avoid randomness
func LocalhostRewriteHook() RecorderHook {
	fn := func(i *cassette.Interaction) error {
		if strings.HasPrefix(i.Request.Host, "localhost") || strings.HasPrefix(i.Request.Host, "127.0.0.1") {
			i.Request.URL = strings.Replace(i.Request.URL, i.Request.Host, DefaultHost, 1)
			i.Request.Host = DefaultHost
		}

		return nil
	}
	return NewRecorderHook(HookNameLocalhostRewrite, fn, recorder.BeforeSaveHook)
}

// FixedDurationHook changes the duration of record HTTP interaction to constant, to avoid randomness
func FixedDurationHook(duration time.Duration) RecorderHook {
	fn := func(i *cassette.Interaction) error {
		i.Response.Duration = duration
		return nil
	}
	return NewRecorderHook(HookNameFixedDuration, fn, recorder.BeforeSaveHook)
}

/************************
	helpers
 ************************/

func mediaType(header http.Header) string {
	v := header.Get("Content-Type")
	media, _, _ := mime.ParseMediaType(v)
	return media
}

func sanitizeValues(values map[string][]string, sanitizers map[string]ValueSanitizer, keys utils.StringSet) map[string][]string {
	for k := range values {
		if !keys.Has(k) {
			continue
		}
		sanitizer, ok := sanitizers[k]
		if !ok {
			sanitizer = DefaultValueSanitizer()
		}
		for i := range values[k] {
			values[k][i] = sanitizer(values[k][i]).(string)
		}
	}
	return values
}

func sanitizeHeaders(headers http.Header, headerKeys utils.StringSet) http.Header {
	return sanitizeValues(headers, HeaderSanitizers, headerKeys)
}

func sanitizeUrl(raw string, queryKeys utils.StringSet) string {
	parsed, e := url.Parse(raw)
	if e != nil {
		return raw
	}
	var queries url.Values = sanitizeValues(parsed.Query(), QuerySanitizers, queryKeys)
	parsed.RawQuery = queries.Encode()
	return parsed.String()
}

func sanitizeRequestForm(req *cassette.Request, queryKeys utils.StringSet) string {
	req.Form = sanitizeValues(req.Form, QuerySanitizers, queryKeys)
	return req.Form.Encode()
}

func sanitizeJsonBody(body string, sanitizers map[string]ValueSanitizer, jsonPaths []parsedJsonPath) string {
	if len(jsonPaths) == 0 {
		return body
	}

	root, e := ajson.Unmarshal([]byte(body))
	if e != nil {
		return body
	}
	for _, path := range jsonPaths {
		nodes, e := ajson.ApplyJSONPath(root, path.Parsed)
		if e != nil || len(nodes) == 0 {
			continue
		}
		for _, node := range nodes {
			sanitizer, ok := sanitizers[node.Key()]
			if !ok {
				sanitizer = DefaultValueSanitizer()
			}
			switch node.Type() {
			case ajson.String:
				_ = node.Set(sanitizer(node.MustString()))
			case ajson.Numeric:
				_ = node.Set(sanitizer(node.MustNumeric()))
			case ajson.Bool:
				_ = node.Set(sanitizer(node.MustBool()))
			default:
			}
		}
	}
	return root.String()
}
