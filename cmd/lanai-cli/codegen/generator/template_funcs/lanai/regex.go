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

package lanai

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"strings"
)

var (
	predefinedRegexes = map[string]string{
		//use go-validator's own value for email regex
		"email":     "",
		"uuid":      "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$",
		"date":      "^\\\\d{4}-\\\\d{2}-\\\\d{2}$",
		"date-time": "^\\\\d{4}-\\\\d{2}-\\\\d{2}T\\\\d{2}:\\\\d{2}:\\\\d{2}(?:\\\\.\\\\d+)?(?:Z|[\\\\+-]\\\\d{2}:\\\\d{2})?$",
	}

	validatedRegexes map[string]string
)

type Regex struct {
	Value string
	Name  string
}

func NewRegex(value openapi3.Schema) (*Regex, error) {
	if !value.Type.Is(openapi3.TypeString) {
		return nil, errors.New("schema is not a string type for regex")
	}

	if value, ok := predefinedRegexes[value.Format]; ok {
		return &Regex{
			Value: value,
			Name:  generateNameFromRegex(value),
		}, nil
	}

	r := Regex{}
	if value.Pattern != "" {
		r.Value = value.Pattern
	} else if value.Format != "" && strings.ToLower(value.Format) != "password" && strings.ToLower(value.Format) != "email" {
		r.Value = value.Format
	} else {
		return nil, nil
	}
	r.Value = strings.ReplaceAll(r.Value, "\\", "\\\\")
	r.Name = generateNameFromRegex(r.Value)
	return &r, nil
}

func generateNameFromRegex(regex string) string {
	for predefinedRegexName, predefinedRegexValue := range predefinedRegexes {
		if predefinedRegexValue == regex {
			return predefinedRegexName
		}
	}

	hashedString := strings.ToUpper(fmt.Sprintf("%x", md5.Sum([]byte(regex))))[0:5]
	return fmt.Sprintf("regex%v", hashedString)
}

func registerRegex(value openapi3.Schema) (string, error) {
	r, err := NewRegex(value)
	if err != nil || (r == nil || r.Value == "") || validatedRegexes[r.Value] != "" {
		return "", err
	}

	validatedRegexes[r.Value] = r.Name
	return r.Name, nil
}
