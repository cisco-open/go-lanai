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

type Response openapi3.ResponseRef

func (r Response) CountFields() (result int) {
	content := Content(r.Value.Content)
	result += content.CountFields()
	return result
}

func (r Response) ContainsRef() bool {
	result := r.Ref != ""
	if !result {
		content := Content(r.Value.Content)
		result = content.ContainsRef()
	}
	return result
}

func (r Response) RefsUsed() (result []string) {
	var refs []string
	if r.Ref != "" {
		// Double check that this isn't an empty wrapper around another ref
		if r.CountFields() != 1 {
			refs = append(refs, path.Base(r.Ref))
		}
	}

	for _, schema := range r.Schemas() {
		if schema.Ref != "" {
			refs = append(refs, path.Base(schema.Ref))
		}
	}

	return refs
}

func (r Response) Schemas() (result []*openapi3.SchemaRef) {
	for _, c := range r.Value.Content {
		result = append(result, SchemaRef(*c.Schema).AllSchemas()...)
	}
	return result
}
