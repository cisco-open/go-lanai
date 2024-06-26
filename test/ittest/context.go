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
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"net/http"
	"net/url"
	"time"
)

type Mode int

const CLIRecordModeFlag = "record-http"

// Recorder states
const (
	// ModeCommandline lets the commandline or the state in TestMain to determine the mode. Default to ModeReplaying
	ModeCommandline Mode = iota
	ModeReplaying
	ModeRecording
)

// DefaultHTTPDuration default duration of recorded HTTP interaction
const DefaultHTTPDuration = 200 * time.Microsecond

var (
	xInteractionIndexHeader = `X-Http-Record-Index`
	xInteractionSeenHeader  = `X-Http-Request-Seen`

	IgnoredRequestHeaders = utils.NewStringSet(xInteractionIndexHeader)

	FuzzyRequestHeaders    = utils.NewStringSet("Authorization")
	FuzzyRequestQueries    = utils.NewStringSet("password", "secret", "nonce", "token", "access_token")
	FuzzyRequestJsonPaths  = utils.NewStringSet()
	FuzzyResponseHeaders   = utils.NewStringSet("Date")
	FuzzyResponseJsonPaths = utils.NewStringSet("$..access_token")
)

/*************************
	HTTPVCROptions
 *************************/

type HTTPVCROptions func(opt *HTTPVCROption)
type HTTPVCROption struct {
	Name               string
	Mode               Mode
	SavePath           string
	RecordMatching     []RecordMatcherOptions
	Hooks              []RecorderHook
	RealTransport      http.RoundTripper
	SkipRequestLatency bool
	// special record matcher that enforce interaction order.
	// to change, use DisableHttpRecordOrdering
	indexAwareWrapper *indexAwareMatcherWrapper
}

/******************************
	HTTP VCR Request Matching
 ******************************/

type GenericMatcherFunc[O, R any] func(O, R) error

type RecordMatcherFunc GenericMatcherFunc[*http.Request, cassette.Request]
type RecordURLMatcherFunc GenericMatcherFunc[*url.URL, *url.URL]
type RecordQueryMatcherFunc GenericMatcherFunc[url.Values, url.Values]
type RecordHeaderMatcherFunc GenericMatcherFunc[http.Header, http.Header]
type RecordBodyMatcherFunc GenericMatcherFunc[[]byte, []byte]
type RecordBodyMatcher interface {
	Support(contentType string) bool
	Matches(out []byte, record []byte) error
}

/******************************
	HTTP VCR Hooks
 ******************************/

// RecorderHook wrapper of recorder.Hook
type RecorderHook interface {
	Name() string
	Handler() recorder.HookFunc
	Kind() recorder.HookKind
}

/******************************
	Request Matcher Logic Ops
 ******************************/

// Note: this is currently not used, we kept it for reference

// AndMatcher generic AND operator of given matchers
// Note: because golang generics requires instantiation, type casting is required.
// 		 e.g. var m RecordBodyMatcherFunc = RecordBodyMatcherFunc(AndMatcher(matcher1, matcher2))
//func AndMatcher[O, R any](matchers ...GenericMatcherFunc[O, R]) GenericMatcherFunc[O, R] {
//	return func(out O, record R) error {
//		for _, matcher := range matchers {
//			if e := matcher(out, record); e != nil {
//				return e
//			}
//		}
//		return nil
//	}
//}

// OrMatcher generic OR operator of given matchers
// Note: because golang generics requires instantiation, type casting is required.
// 		 e.g. var m RecordBodyMatcherFunc = RecordBodyMatcherFunc(AndMatcher(matcher1, matcher2))
//func OrMatcher[O, R any](matchers ...GenericMatcherFunc[O, R]) GenericMatcherFunc[O, R] {
//	return func(out O, record R) error {
//		var e error
//		for _, matcher := range matchers {
//			if e = matcher(out, record); e == nil {
//				return nil
//			}
//		}
//		return e
//	}
//}
