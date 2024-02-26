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
    "fmt"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/util"
    "github.com/getkin/kin-openapi/openapi3"
)

func requiredList(val interface{}) ([]string, error) {
	var list []string
	switch val.(type) {
	case *openapi3.SchemaRef:
		list = val.(*openapi3.SchemaRef).Value.Required
	case *openapi3.Parameter:
		parameter := val.(*openapi3.Parameter)
		if parameter.Required {
			list = append(list, parameter.Name)
		}
	case *openapi3.RequestBody:
		reqBody := val.(*openapi3.RequestBody)
		if reqBody.Required {
			list = append(list, "Body")
		}
	default:
		return nil, fmt.Errorf("requiredList error: unsupported interface %v", util.GetInterfaceType(val))
	}
	return list, nil
}
