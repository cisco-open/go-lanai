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

//Representations of common structs in the templates, letting them look a little cleaner

import (
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
	"github.com/getkin/kin-openapi/openapi3"
	"text/template"
)

var (
	propertyFuncMap = template.FuncMap{}
)

type Schema struct {
	Name string
	Data *openapi3.SchemaRef
}

func NewSchema(name string, data *openapi3.SchemaRef) Schema {
	return Schema{
		Name: name,
		Data: data,
	}
}

func (s Schema) AllSchemas() openapi3.SchemaRefs {
	return SchemaRef(*s.Data).AllSchemas()
}
func (s Schema) StructProperties() (result openapi3.SchemaRefs) {
	return SchemaRef(*s.Data).StructProperties()
}

func (s Schema) AllProperties() (result openapi3.Schemas) {
	result = make(openapi3.Schemas)
	if s.Data.Value.Type == nil || s.Data.Value.Type.Is(openapi3.TypeObject) {
		result = s.Data.Value.Properties
	}
	return result
}

func (s Schema) Type() string {
	if s.Data.Value.Type == nil || len(*s.Data.Value.Type) == 0 {
		return openapi3.TypeObject
	}
	return (*s.Data.Value.Type)[0]
}
func (s Schema) HasAdditionalProperties() bool {
	return openapi.SchemaRef(*s.Data).HasAdditionalProperties()
}
