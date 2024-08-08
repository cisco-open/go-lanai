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

package lanaiutil

import (
    "fmt"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/util"
    "github.com/getkin/kin-openapi/openapi3"
    "reflect"
    "strings"
)

const (
	UUID_IMPORT_PATH = "github.com/google/uuid"
	JSON_IMPORT_PATH = "encoding/json"
	TIME_IMPORT_PATH = "time"
)

var (
	FormatToExternalImport = map[string]string{
		"uuid":      UUID_IMPORT_PATH,
		"date-time": TIME_IMPORT_PATH,
	}
)

func ExternalImportsFromFormat(element interface{}) (result []string) {
	for format, externalImport := range FormatToExternalImport {
		if MatchesFormat(element, format) {
			result = append(result, externalImport)
		}
	}
	return
}

func MatchesFormat(element interface{}, specificType string) bool {
	schema, err := ConvertToSchemaRef(element)
	if err != nil || schema == nil || schema.Value.Type == nil || !schema.Value.Type.Is(openapi3.TypeString)  {
		return false
	}

	formatMatchesType := strings.ToLower(schema.Value.Pattern) == specificType || strings.ToLower(schema.Value.Format) == specificType
	// exclude path parameters because go's validation only supports base types, so this should stay as a string
	isNotInPathParameter := reflect.TypeOf(element) != reflect.TypeOf(&openapi3.Parameter{}) || element.(*openapi3.Parameter).In != "path"
	isNotInQueryParameter := reflect.TypeOf(element) != reflect.TypeOf(&openapi3.Parameter{}) || element.(*openapi3.Parameter).In != "query"

	return formatMatchesType && (isNotInPathParameter && isNotInQueryParameter)
}

func ConvertToSchemaRef(element interface{}) (*openapi3.SchemaRef, error) {
	var val *openapi3.SchemaRef
	switch v := element.(type) {
	case *openapi3.SchemaRef:
		val = v
	case *openapi3.Parameter:
		val = v.Schema
	case openapi3.AdditionalProperties:
		val = v.Schema
	default:
		return nil, fmt.Errorf("ConvertToSchemaRef: unsupported interface %v", util.GetInterfaceType(element))
	}
	return val, nil
}
