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
    "errors"
    _go "github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/go"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
    "github.com/getkin/kin-openapi/openapi3"
    "path"
    "strings"
)

type Operation struct {
	Name string
	Data *openapi3.Operation
}

func NewOperation(data *openapi3.Operation, defaultName string) Operation {
	name := defaultName
	if data.OperationID != "" {
		name = data.OperationID
	}
	return Operation{
		Name: name,
		Data: data,
	}
}

func (o Operation) StructForMessage(messageType string, structRegistry map[string]string) (*_go.Struct, error) {
	switch strings.ToLower(messageType) {
	case "request":
		return o.RequestStruct(structRegistry), nil
	case "response":
		return o.ResponseStruct(structRegistry), nil
	default:
		return nil, errors.New("type must be \"request\" or \"response\"")
	}
}

func (o Operation) RequestRefsUsed() (result []string) {
	for _, p := range o.Data.Parameters {
		result = append(result, p.Ref)
	}
	if o.Data.RequestBody != nil {
		r := openapi.RequestBody(*o.Data.RequestBody)
		result = append(result, r.RefsUsed()...)
	}

	return result
}

func (o Operation) ResponseRefsUsed() (result []string) {
	responses := NewResponses(o.Data.Responses).Sorted()
	for _, resp := range responses {
		if resp.CountFields() == 1 && resp.ContainsRef() {
			result = append(result, resp.RefsUsed()...)
			break
		}
		break
	}
	return result
}

func (o Operation) RequestStruct(structRegistry map[string]string) *_go.Struct {
	structName := o.Name + "Request"
	var structPackage string
	p, ok := structRegistry[strings.ToLower(structName)]
	if ok {
		structPackage = p
	} else {
		refs := o.RequestRefsUsed()
		if refs == nil {
			return nil
		}
		singularRef := refs[0]
		structName = path.Base(singularRef)

		p, ok := structRegistry[strings.ToLower(structName)]
		if ok {
			structPackage = p
		}
	}
	return &_go.Struct{
		Package: structPackage,
		Name:    structName,
	}
}

func (o Operation) ResponseStruct(structRegistry map[string]string) *_go.Struct {
	structName := o.Name + "Response"
	structPackage := structRegistry[strings.ToLower(structName)]
	if structPackage == "" {
		refsUsed := o.ResponseRefsUsed()
		if len(refsUsed) == 0 {
			return nil
		}
		responseRef := refsUsed[0]
		structName = path.Base(responseRef)
		p, ok := structRegistry[strings.ToLower(structName)]
		if ok {
			structPackage = p
		}
	}
	return &_go.Struct{
		Package: structPackage,
		Name:    structName,
	}
}

func (o Operation) AllResponseContent() (result []*openapi3.MediaType) {
	responses := NewResponses(o.Data.Responses).Sorted()
	for _, response := range responses {
		for _, content := range response.Value.Content {
			result = append(result, content)
		}
	}
	return result
}
