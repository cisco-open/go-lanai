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
	_go "cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/go"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
	"github.com/getkin/kin-openapi/openapi3"
)

func ImportsUsedByPath(pathItem openapi3.PathItem, repositoryPath string) []string {
	var allImports []string
	for _, operation := range pathItem.Operations() {
		responses := Responses{openapi.Responses(operation.Responses)}
		parameters := Parameters{openapi.Parameters(operation.Parameters)}
		var requestBody RequestBody
		if operation.RequestBody != nil {
			requestBody = RequestBody{openapi.RequestBody(*operation.RequestBody)}
		}
		refs := responses.RefsUsed()
		numFieldsInRequestStruct := parameters.CountFields() + requestBody.CountFields()
		if numFieldsInRequestStruct != 1 {
			refs = append(refs, parameters.RefsUsed()...)
			refs = append(refs, requestBody.RefsUsed()...)
		}

		for _, i := range getImportsFromRef(refs) {
			allImports = append(allImports, repositoryPath+"/"+i)
		}

		// Grab any external dependencies
		allImports = append(allImports, responses.ExternalImports()...)
		allImports = append(allImports, parameters.ExternalImports()...)
		allImports = append(allImports, requestBody.ExternalImports()...)

	}

	uniqueImports := make(map[string]bool)
	for _, r := range allImports {
		uniqueImports[r] = true
	}

	var result []string
	for k := range uniqueImports {
		result = append(result, k)
	}
	return result
}

func getImportsFromRef(refs []string) []string {
	var imports []string
	for _, ref := range refs {
		loc := _go.StructLocation(ref)
		if loc != "" {
			imports = append(imports, loc)
		}
	}
	return imports
}
