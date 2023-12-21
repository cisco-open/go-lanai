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

package gomegautils

import (
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/spyzhov/ajson"
	"strings"
)

/****************************
	Common Gomega Matchers
 ****************************/

// HaveJsonPathWithValue returns a gomega matcher.
// This matcher extract fields from JSON string using JSONPath, and assert that if the result slice matches the expected value
//
// "value" can be following types:
// - types.GomegaMatcher, then the given matcher is directly applied to the slice resulted from JSONPath searching
// - any non-matcher type, HaveJsonPathWithValue by default use gomega.ContainElements(gomega.Equal(expected)) on any non-matcher value
//
// Following statements are equivalent:
// 		Expect(jsonStr).To(HaveJsonPathWithValue("$..status", "GOOD"))
//		Expect(jsonStr).To(HaveJsonPathWithValue("$..status", gomega.ContainElements(gomega.Equal("GOOD"))))
func HaveJsonPathWithValue(jsonPath string, value interface{}) types.GomegaMatcher {
	var matcher types.GomegaMatcher
	switch v := value.(type) {
	case types.GomegaMatcher:
		matcher = v
	default:
		matcher = gomega.ContainElements(gomega.Equal(v))
	}
	return &GomegaJsonPathMatcher{
		jsonPath:       jsonPath,
		delegate:       matcher,
	}
}

// HaveJsonPath returns a gomega matcher, similar to HaveJsonPathWithValue
// HaveJsonPath succeed only if the specified JSONPath yield non-empty result from actual JSON string.
//
// Following statements are equivalent:
// 		Expect(jsonStr).To(HaveJsonPath("$..status"))
//		Expect(jsonStr).To(HaveJsonPath("$..status", gomega.Not(gomega.BeEmpty)))
func HaveJsonPath(jsonPath string) types.GomegaMatcher {
	return &GomegaJsonPathMatcher{
		jsonPath:       jsonPath,
		delegate:       Not(BeEmpty()),
	}
}

type GomegaJsonPathMatcher struct {
	jsonPath       string
	delegate       types.GomegaMatcher
}

func (m *GomegaJsonPathMatcher) Match(actual interface{}) (success bool, err error) {
	values, e := m.jsonPathValues(actual)
	if e != nil {
		return false, e
	}
	return m.delegate.Match(values)
}

func (m *GomegaJsonPathMatcher) FailureMessage(actual interface{}) (message string) {
	msg := fmt.Sprintf("to have JsonPath %s matching", m.jsonPath)
	desc := format.Message(strings.TrimSpace(asString(actual)), msg, m.delegate)
	actual, _ = m.jsonPathValues(actual)
	return fmt.Sprintf("%s\nResult:\n%s", desc, m.delegate.FailureMessage(actual))
}

func (m *GomegaJsonPathMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	msg := fmt.Sprintf("to have JsonPath %s not matching", m.jsonPath)
	desc := format.Message(strings.TrimSpace(asString(actual)), msg, m.delegate)
	actual, _ = m.jsonPathValues(actual)
	return fmt.Sprintf("%s\nResult:\n%s", desc, m.delegate.NegatedFailureMessage(actual))
}

func (m *GomegaJsonPathMatcher) jsonPathValues(actual interface{}) ([]interface{}, error) {
	var data []byte
	switch v := actual.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return nil, fmt.Errorf("expect string or []byte, but got %T", actual)
	}

	root, e := ajson.Unmarshal(data)
	if e != nil {
		return nil, fmt.Errorf(`expect json string but got %T`, actual)
	}
	parsed, e := ajson.ParseJSONPath(m.jsonPath)
	if e != nil {
		return nil, fmt.Errorf("invalid JSONPath '%s'", m.jsonPath)
	}
	nodes, e := ajson.ApplyJSONPath(root, parsed)
	if e != nil {
		return nil, fmt.Errorf(`invalid JsonPath "%s"`, m.jsonPath)
	}
	values := make([]interface{}, len(nodes))
	for i, node := range nodes {
		var e error
		if values[i], e = node.Unpack(); e != nil {
			return nil, fmt.Errorf(`unable to extract value of JsonPath [%s]: %v'`, m.jsonPath, e)
		}
	}
	return values, nil
}

/****************************
	Gomega Matchers Helpers
 ****************************/

func asString(actual interface{}) string {
	var data string
	switch v := actual.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	}
	return data
}

