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

type SchemaRef openapi3.SchemaRef

// AllSchemas will return every schemaRef object associated with this object
// Will Include:
// - The base schema
// - All the schemas part of the AllOf array (recursively)
// - Items, if the schema is an array (recursively)
func (s SchemaRef) AllSchemas() (result openapi3.SchemaRefs) {
	base := openapi3.SchemaRef(s)
	result = append(result, &base)
	for _, a := range s.Value.AllOf {
		result = append(result, SchemaRef(*a).AllSchemas()...)
	}
	if s.Value.Type == "array" {
		result = append(result, SchemaRef(*s.Value.Items).AllSchemas()...)
	}

	return
}

// HasAdditionalProperties returns true if the schema supports additionalProperties of any kind
func (s SchemaRef) HasAdditionalProperties() bool {
	additionalPropertiesAllowed := s.Value.AdditionalProperties.Has != nil && *s.Value.AdditionalProperties.Has
	additionalPropsDefined := s.Value.AdditionalProperties.Schema != nil

	return additionalPropertiesAllowed || additionalPropsDefined
}
