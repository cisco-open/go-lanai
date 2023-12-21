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

package utils

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode"
)

func UnQuote(s string) string {
	if len(s) >= 2 {
		if c := s[len(s)-1]; s[0] == c && (c == '"' || c == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func ParseString(s string) interface{} {
	// try number before boolean because 1/0 can be parsed to boolean true/false
	if numVal, err := strconv.ParseFloat(s, 64); err == nil {
		return numVal
	} else if boolVal, err := strconv.ParseBool(s); err == nil {
		return boolVal
	}

	// we also support []interface{} and map[string]interface{}
	trimmed := strings.TrimSpace(s)
	if strings.HasPrefix(trimmed, "{") {
		var v map[string]interface{}
		if e := json.Unmarshal([]byte(trimmed), &v); e == nil {
			return v
		}
	}

	if strings.HasPrefix(trimmed, "[") {
		var v []interface{}
		if e := json.Unmarshal([]byte(trimmed), &v); e == nil {
			return v
		}
	}

	return s
}

const dash = rune('-')

// CamelToSnakeCase convert "camelCase" string to "snake-case"
func CamelToSnakeCase(camelCase string) string {
	var converted []rune
	for pos, char := range camelCase {
		if unicode.IsUpper(char) {
			if pos>0 && unicode.IsLower([]rune(camelCase)[pos-1]) {
				converted = append(converted, dash)
			}
			converted = append(converted, unicode.ToLower(char))
		} else {
			converted = append(converted, unicode.ToLower(char))
		}
	}
	return string(converted)
}
