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

type Components openapi3.Components

func (c Components) AllProperties() (result []openapi3.Schemas) {
	var schemas openapi3.SchemaRefs
	for _, schema := range c.Schemas {
		s := SchemaRef(*schema)
		schemas = append(schemas, s.AllSchemas()...)
	}

	for _, requestBody := range c.RequestBodies {
		if requestBody.Value == nil {
			continue
		}
		schemas = append(schemas, RequestBody(*requestBody).Schemas()...)
	}

	schemas = append(schemas, FromParameterMap(c.Parameters).schemas()...)
	for _, s := range schemas {
		// If there is a Ref, that doesn't count as a property
		if s.Ref == "" {
			result = append(result, s.Value.Properties)
		}
	}
	return
}
