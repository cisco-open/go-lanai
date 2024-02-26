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

type Responses struct {
	openapi.Responses
}

func NewResponses(src *openapi3.Responses) *Responses {
	if src == nil {
		return &Responses{}
	}
	return &Responses{Responses: openapi.Responses{Responses: *src}}
}

func (r Responses) ExternalImports() (result []string) {
	for _, response := range r.Responses.Map() {
		resp := Response{openapi.Response(*response)}
		// check if a struct will be created from this response
		if resp.CountFields() == 0 || (resp.CountFields() == 1 && resp.ContainsRef()) {
			break
		}
		result = append(result, resp.ExternalImports()...)
	}
	return result
}

type Response struct {
	openapi.Response
}

func (r Response) ExternalImports() (result []string) {
	if r.Ref != "" {
		return result
	}
	for _, schema := range r.Schemas() {
		result = append(result, lanaiutil.ExternalImportsFromFormat(schema)...)
	}
	return result
}
