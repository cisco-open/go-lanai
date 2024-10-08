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

package openapi

import "github.com/getkin/kin-openapi/openapi3"

type Content openapi3.Content

func (c *Content) ContainsRef() bool {
	for _, c := range *c {
		if c.Schema.Ref != "" {
			return true
		} else {
			if c.Schema.Value.Type.Is(openapi3.TypeArray) {
				return c.Schema.Value.Items.Ref != ""
			}
		}
	}
	return false
}

func (c *Content) CountFields() (result int) {
	for _, content := range *c {
		if content.Schema.Ref != "" {
			result++
		} else {
			if content.Schema.Value.Type.Is(openapi3.TypeObject) {
				result++
			}
			result += len(content.Schema.Value.AllOf)
		}
	}
	return result
}
