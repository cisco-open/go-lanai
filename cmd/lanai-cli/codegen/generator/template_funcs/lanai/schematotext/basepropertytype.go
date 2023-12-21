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

package schematotext

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
)

type baseType struct {
	data interface{}
}

func NewBaseType(data interface{}, opts ...func(option *translatorOptions)) *baseType {
	o := &translatorOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return &baseType{
		data: data,
	}
}

func (b baseType) ToText() (string, error) {
	schema, err := lanaiutil.ConvertToSchemaRef(b.data)
	if err != nil {
		return "", err
	}
	return schemaToGoBaseTypes(schema.Value).String(), nil
}
func schemaToGoBaseTypes(val *openapi3.Schema) (result reflect.Type) {
	switch val.Type {
	case openapi3.TypeBoolean:
		result = reflect.TypeOf(true)
	case openapi3.TypeNumber:
		result = reflect.TypeOf(1.1)
	case openapi3.TypeInteger:
		var v interface{}
		switch val.Format {
		case "int32":
			v = int32(1)
		case "int64":
			v = int64(1)
		default:
			v = 1
		}
		result = reflect.TypeOf(v)
	case openapi3.TypeString:
		result = reflect.TypeOf("string")
	default:
		result = nil
	}
	return
}
