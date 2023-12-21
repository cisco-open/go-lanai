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
	"fmt"
	"github.com/spyzhov/ajson"
	"mime"
	"net/url"
	"reflect"
	"strings"
)

type RecordLiteralBodyMatcher GenericMatcherFunc[[]byte, []byte]

func (m RecordLiteralBodyMatcher) Support(_ string) bool {
	return true
}

func (m RecordLiteralBodyMatcher) Matches(out []byte, record []byte) error {
	return m(out, record)
}

// NewRecordLiteralBodyMatcher  returns RecordBodyMatcher that matches request bodies literally
func NewRecordLiteralBodyMatcher() RecordBodyMatcher {
	return RecordLiteralBodyMatcher(func(out []byte, record []byte) error {
		if len(out) != len(record) {
			return fmt.Errorf("body lengths mismatch")
		}
		for i := range record {
			if out[i] != record[i] {
				return fmt.Errorf("body content mismatch")
			}
		}
		return nil
	})
}

type RecordFormBodyMatcher GenericMatcherFunc[[]byte, []byte]

func (m RecordFormBodyMatcher) Support(contentType string) bool {
	media, _, e := mime.ParseMediaType(contentType)
	return e == nil && media == "application/x-www-form-urlencoded"
}

func (m RecordFormBodyMatcher) Matches(out []byte, record []byte) error {
	return m(out, record)
}

// NewRecordFormBodyMatcher returns RecordBodyMatcher that matches request bodies as application/x-www-form-urlencoded.
// any value in the fuzzyKeys is not compared. But outgoing body need to have all keys contained in the record body
func NewRecordFormBodyMatcher(fuzzyKeys...string) RecordBodyMatcher {
	valuesMatcher := newValuesMatcher("form body", nil, fuzzyKeys...)
	return RecordLiteralBodyMatcher(func(out []byte, record []byte) error {
		outForm := parseFormBody(out)
		rForm := parseFormBody(record)
		return valuesMatcher(outForm, rForm)
	})
}

type RecordJsonBodyMatcher GenericMatcherFunc[[]byte, []byte]

func (m RecordJsonBodyMatcher) Support(contentType string) bool {
	media, _, e := mime.ParseMediaType(contentType)
	return e == nil && media == "application/json"
}

func (m RecordJsonBodyMatcher) Matches(out []byte, record []byte) error {
	return m(out, record)
}

// NewRecordJsonBodyMatcher returns a RecordBodyMatcher that matches JSON body of recorded and outgoing request.
// Values of any field matching the optional fuzzyJsonPaths is not compared, but outgoing request body must contain
// all fields that the record contains
func NewRecordJsonBodyMatcher(fuzzyJsonPaths ...string) RecordBodyMatcher {
	parsedPaths := parseJsonPaths(fuzzyJsonPaths)
	return RecordJsonBodyMatcher(func(out []byte, record []byte) error {
		rRoot, rMatched, e := parseJsonWithFilter(record, parsedPaths)
		if e != nil {
			return e
		}
		lRoot, lMatched, e := parseJsonWithFilter(out, parsedPaths)
		if e != nil {
			return e
		}

		// first compare filtered nodes, they need to be identical
		if !reflect.DeepEqual(lRoot, rRoot) {
			return fmt.Errorf("JSON body content mismatch")
		}

		// second, check if all matched fuzzy json paths in record exists in the outgoing body
	OUTER:
		for _, p := range rMatched {
			for _, lp := range lMatched {
				if p.Value == lp.Value {
					continue OUTER
				}
			}
			return fmt.Errorf("JSON body content mismatch: missing [%s]", p.Value)
		}
		return nil
	})
}

/**********************
	helpers
 **********************/

type parsedJsonPath struct {
	Value  string
	Parsed []string
}

func parseJsonPaths(jsonPaths []string) (parsed []parsedJsonPath) {
	parsed = make([]parsedJsonPath, 0, len(jsonPaths))
	for _, path := range jsonPaths {
		p, e := ajson.ParseJSONPath(path)
		if e != nil {
			panic(e)
		}
		parsed = append(parsed, parsedJsonPath{Value: path, Parsed: p})
	}
	return
}

func parseJsonWithFilter(data []byte, jsonPaths []parsedJsonPath) (filtered interface{}, filteredPaths []parsedJsonPath, err error) {
	root, e := ajson.Unmarshal(data)
	if e != nil {
		return nil, nil, e
	}
	filteredPaths = make([]parsedJsonPath, 0, len(jsonPaths))
	for _, path := range jsonPaths {
		nodes, e := ajson.ApplyJSONPath(root, path.Parsed)
		if e != nil || len(nodes) == 0 {
			continue
		}
		filteredPaths = append(filteredPaths, path)
		for _, node := range nodes {
			_ = node.Delete()
		}
	}
	filtered, err = root.Unpack()
	return
}

func parseFormBody(data []byte) url.Values {
	parsed := url.Values{}
	vals := strings.Split(string(data), "&")
	for _, pair := range vals {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) < 2 {
			continue
		}
		if v , e := url.QueryUnescape(kv[1]); e == nil {
			parsed.Add(kv[0], v)
		}
	}
	return parsed
}

