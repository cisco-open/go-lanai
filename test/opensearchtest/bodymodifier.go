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

package opensearchtest

import (
	"github.com/tidwall/sjson"
	"testing"
)

// MatcherBodyModifiers provides a way to control the MatcherBodyModifier that is
// passed into the MatchBody function.
// Deprecated: Use FuzzyJsonPaths instead
type MatcherBodyModifiers []MatcherBodyModifier

// Modifier will return pointer to the slice of MatcherBodyModifier
func (m *MatcherBodyModifiers) Modifier() []MatcherBodyModifier {
	return *m
}

// Append can be used to append a new MatcherBodyModifier
func (m *MatcherBodyModifiers) Append(modifier MatcherBodyModifier) {
	*m = append(*m, modifier)
}

// Clear is used to clear all of the existing MatcherBodyModifier
func (m *MatcherBodyModifiers) Clear() {
	*m = nil
}

// MatcherBodyModifier will modify the body of a request that goes to the MatchBody
// to remove things that might make matching difficult.
// Example being time parameters in queries, or randomly generated values.
// To see this in use, check out SubTestTimeBasedQuery in opensearch_test.go
// Deprecated: Use FuzzyJsonPaths instead
type MatcherBodyModifier func(*[]byte)

// IgnoreGJSONPaths will ignore any of the fields that are defined by the gjsonPaths
// which follow the GJSON syntax.
// https://github.com/tidwall/gjson/blob/master/SYNTAX.md#gjson-path-syntax
// Deprecated: Use FuzzyJsonPaths instead
func IgnoreGJSONPaths(t *testing.T, gjsonPaths ...string) MatcherBodyModifier {
	return func(b *[]byte) {
		var err error
		for _, path := range gjsonPaths {
			*b, err = sjson.DeleteBytes(*b, path)
			if err != nil {
				t.Errorf("unable to delete bytes: %v", err)
			}
		}
	}
}
