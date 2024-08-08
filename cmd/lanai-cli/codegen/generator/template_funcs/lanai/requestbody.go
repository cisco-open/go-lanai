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
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
	"github.com/getkin/kin-openapi/openapi3"
)

type RequestBody struct {
	openapi.RequestBody
}

func NewRequestBody(body openapi3.RequestBodyRef) RequestBody {
	return RequestBody{openapi.RequestBody(body)}
}

func (r RequestBody) ExternalImports() (result []string) {
	if r.CountFields() == 0 {
		return
	}
	if r.Value == nil {
		return
	}
	for _, schema := range r.Schemas() {
		result = append(result, lanaiutil.ExternalImportsFromFormat(schema)...)
		if schema.Ref == "" {
			if schema.Value != nil && schema.Value.Type != nil && !schema.Value.Type.Is(openapi3.TypeObject) {
				result = append(result, lanaiutil.JSON_IMPORT_PATH)
			}
		}
	}
	return result
}
