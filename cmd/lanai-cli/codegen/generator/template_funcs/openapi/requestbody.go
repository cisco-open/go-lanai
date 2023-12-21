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

import (
	"github.com/getkin/kin-openapi/openapi3"
	"path"
)

type RequestBody openapi3.RequestBodyRef

func (r RequestBody) ContainsRef() (result bool) {
	if r.Ref != "" {
		return true
	}
	if r.Value == nil {
		return false
	}
	for _, j := range r.Value.Content {
		if j.Schema.Ref != "" {
			result = true
		}
	}
	return result
}

func (r RequestBody) CountFields() (result int) {
	if r.Ref != "" {
		result++
	}
	if r.Value != nil {
		result += len(r.Value.Content)
	}
	return result
}

func (r RequestBody) RefsUsed() (result []string) {
	if r.CountFields() == 0 {
		return
	}
	if r.Ref != "" {
		result = append(result, path.Base(r.Ref))
	}

	if r.Value == nil {
		return
	}
	//Assumption - Responses will have just one mediatype defined in contract, e.g just "application/json"
	if len(r.Value.Content) > 1 {
		logger.Warn("more than one mediatype defined in requestBody, undefined behavior may occur")
	}
	for _, schema := range r.Schemas() {
		if schema.Ref != "" {
			result = append(result, path.Base(schema.Ref))
		} else if schema.Value.Type == "array" && schema.Value.Items.Ref != "" {
			result = append(result, schema.Value.Items.Ref)
		}
	}
	return result
}

func (r RequestBody) Schemas() (result openapi3.SchemaRefs) {
	for _, c := range r.Value.Content {
		result = append(result, SchemaRef(*c.Schema).AllSchemas()...)
	}
	return result
}
