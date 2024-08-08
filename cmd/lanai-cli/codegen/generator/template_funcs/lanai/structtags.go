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
    "fmt"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/util"
    "github.com/getkin/kin-openapi/openapi3"
    "strings"
)

func structTags(p Property) string {
	result := ""

	nameType := nameType(p.PropertyData)
	name := p.PropertyName
	if p.OmitJSON {
		name = "-"
	} else if nameType == "json" {
		schema, _ := lanaiutil.ConvertToSchemaRef(p.PropertyData)
		if p.IsOptional() && !schema.Value.Type.Is(openapi3.TypeArray) && (ShouldHavePointer(p) || !zeroValueIsValid(schema)) {
			name = fmt.Sprintf("%v,omitempty", name)
		}
	}

	result = result + fmt.Sprintf("%v:\"%v%v\"", nameType, name, defaultValue(p))
	binding := bindings(p.PropertyName, p.PropertyData, p.RequiredList)
	if binding != "" {
		result = fmt.Sprintf("%v binding:\"%v\"", result, binding)
	}

	result = fmt.Sprintf("`%v`", result)
	return result
}

func defaultValue(p Property) string {
	// default values don't make sense for required properties
	if !p.IsOptional() {
		return ""
	}

	result := ""
	switch p.PropertyData.(type) {
	case *openapi3.Parameter:
		in := p.PropertyData.(*openapi3.Parameter).In
		if in == "header" || in == "query" {
			schema, _ := lanaiutil.ConvertToSchemaRef(p.PropertyData)
			if schema.Value.Default != nil {
				result = ",default=" + fmt.Sprintf("%v", schema.Value.Default)
			}
		}
	}
	return result
}

func nameType(element interface{}) string {
	nameType := "json"

	switch element.(type) {
	case *openapi3.SchemaRef:
		nameType = "json"
	case *openapi3.Parameter:
		switch element.(*openapi3.Parameter).In {
		case "header":
			nameType = "header"
		case "query":
			nameType = "form"
		case "path":
			fallthrough
		default:
			nameType = "uri"
		}
	default:
		logger.Errorf("no supported added for struct-tags for %v, double check contract or log a bug", util.GetInterfaceType(element))
	}
	return nameType
}

func bindings(propertyName string, data interface{}, requiredList []string) string {
	bindingParts := requiredTag(propertyName, requiredList)
	if isBaseType(data) {
		validationTags := validationTags(data)
		bindingParts = append(bindingParts, omitEmptyTags(propertyName, requiredList, len(validationTags))...)
		bindingParts = append(bindingParts, validationTags...)
	}
	return strings.Join(bindingParts, ",")
}

func validationTags(data interface{}) []string {
	result := regexTag(data)
	schemaRef, err := lanaiutil.ConvertToSchemaRef(data)
	if err != nil {
		return nil
	}
	result = append(result, limitTags(schemaRef)...)
	result = append(result, enumOf(schemaRef.Value.Enum)...)
	if schemaRef.Value.Type.Is(openapi3.TypeArray) {
		innerParts := validationTags(schemaRef.Value.Items)
		if innerParts != nil {
			result = append(result, "dive")
			result = append(result, innerParts...)
		}
	}
	return result
}

func requiredTag(propertyName string, requiredList []string) (result []string) {
	if util.ListContains(requiredList, propertyName) {
		result = append(result, "required")
	}
	return result
}

func regexTag(element interface{}) (result []string) {
	schemaRef, err := lanaiutil.ConvertToSchemaRef(element)
	if err != nil {
		return nil
	}
	rValue, _ := NewRegex(*schemaRef.Value)
	if rValue != nil {
		result = append(result, rValue.Name)
	}
	return result
}

func limitTags(schemaRef *openapi3.SchemaRef) (result []string) {
	if schemaRef == nil {
		return result
	}

	min, max := limitsForSchema(schemaRef.Value)
	if min != "" {
		result = append(result, fmt.Sprintf("min=%v", min))
	}
	if max != "" {
		result = append(result, fmt.Sprintf("max=%v", max))
	}
	return result
}

// omitEmptyTags will adds omitEmpty tag if:
// the property is not in the list of required properties
// numberOfValidationTags > 0 - if there are any validations that need to be omitted
func omitEmptyTags(propertyName string, requiredList []string, numberOfValidationTags int) (result []string) {
	if !util.ListContains(requiredList, propertyName) && numberOfValidationTags > 0 {
		result = append(result, "omitempty")
	}
	return result
}

func enumOf(enums []interface{}) (result []string) {
	allEnums := ""
	for _, e := range enums {
		if e == nil {
			continue
		}
		allEnums += fmt.Sprintf("%v ", e.(string))
	}

	if allEnums == "" {
		return result
	}
	binding := strings.TrimSpace("enumof=" + allEnums)
	result = append(result, binding)
	return result
}

func isBaseType(element interface{}) bool {
	usesExternalType := false
	for format := range lanaiutil.FormatToExternalImport {
		if lanaiutil.MatchesFormat(element, format) {
			usesExternalType = true
			break
		}
	}

	return !usesExternalType
}
