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
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/util"

	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Property struct {
	PropertyName string
	CurrentPkg   string
	TypePrefix   string
	OmitJSON     bool
	RequiredList []string
	PropertyData interface{}
}

func NewProperty(data interface{}, name string, requiredList []string, currentPkg string, typePrefix ...string) Property {
	return Property{
		PropertyName: name,
		PropertyData: data,
		CurrentPkg:   currentPkg,
		RequiredList: requiredList,
		TypePrefix:   strings.Join(typePrefix, ""),
	}
}

func (p Property) SetOmitJSON(val bool) Property {
	p.OmitJSON = val
	return p
}

func (p Property) IsOptional() bool {
	return !util.ListContains(p.RequiredList, p.PropertyName)
}

func PropertyToGoType(p Property) (string, error) {
	defaultObjectName := util.ToTitle(p.TypePrefix + util.ToTitle(p.PropertyName))
	result, err := SchemaToText(p.PropertyData, defaultObjectName, p.CurrentPkg)
	if err != nil {
		return "", err
	}
	return result, nil
}

func ShouldHavePointer(p Property) bool {
	isRequired := util.ListContains(p.RequiredList, p.PropertyName)
	schema, _ := lanaiutil.ConvertToSchemaRef(p.PropertyData)
	if schema.Value.Enum != nil {
		return false
	}
	if len(schema.Value.Properties) == 0 {
		if SchemaRef(*schema).HasAdditionalProperties() {
			return false
		}
	}
	if schema.Value.Nullable {
		return true
	}
	switch schema.Value.Type {
	case openapi3.TypeObject:
		return isRequired
	case openapi3.TypeArray, openapi3.TypeBoolean:
		return false
	default:
		if zeroValueIsValid(schema) {
			return isRequired
		} else {
			return !isRequired
		}
	}
}

func zeroValueIsValid(schema *openapi3.SchemaRef) bool {
	return valuePassesValidation(schema.Value, zeroValue(schema.Value))
}

func zeroValue(schema *openapi3.Schema) reflect.Value {
	goType := schemaToGoType(schema)
	if goType == nil {
		return reflect.Value{}
	}
	zvValue := reflect.Zero(goType)
	return zvValue
}
func schemaToGoType(val *openapi3.Schema) (result reflect.Type) {
	switch val.Type {
	case openapi3.TypeBoolean, openapi3.TypeNumber, openapi3.TypeInteger, openapi3.TypeString:
		result = schemaToGoBaseTypes(val)
	case openapi3.TypeArray:
		itemsType := schemaToGoType(val.Items.Value)
		if itemsType != nil {
			result = reflect.SliceOf(itemsType)
		}
	case openapi3.TypeObject:
	//	Do nothing
	default:
		logger.Warnf("getType: type %v doesn't have corresponding mapping", val.Type)
	}

	return result
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

func valuePassesValidation(schema *openapi3.Schema, value reflect.Value) (result bool) {
	switch value.Kind() {
	case reflect.String:
		if rValue, _ := NewRegex(*schema); rValue != nil {
			found, err := regexp.MatchString(rValue.Value, value.String())
			if err != nil || !found {
				return false
			}
		}
	}
	return valueIsWithinSchemaLimits(schema, value)
}

func valueIsWithinSchemaLimits(schema *openapi3.Schema, value reflect.Value) bool {
	min, max := limitsForSchema(schema)
	switch value.Kind() {
	case reflect.String:
		return !isOutOfBounds(len(value.String()), min, max)
	case reflect.Int:
		return !isOutOfBounds(value.Interface().(int), min, max)
	case reflect.Int32:
		return !isOutOfBounds(value.Interface().(int32), min, max)
	case reflect.Int64:
		return !isOutOfBounds(value.Interface().(int64), min, max)
	case reflect.Float64:
		return !isOutOfBounds(value.Float(), min, max)
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		return !isOutOfBounds(value.Len(), min, max)
	}

	return false
}

func limitsForSchema(element *openapi3.Schema) (min, max string) {
	switch element.Type {
	case "array":
		if element.MinItems > 0 {
			min = strconv.FormatUint(element.MinItems, 10)
		}
		if element.MaxItems != nil {
			max = strconv.FormatUint(*element.MaxItems, 10)
		}
	case "number":
		fallthrough
	case "integer":
		if element.Min != nil {
			min = strconv.FormatFloat(*element.Min, 'f', -1, 64)
		}
		if element.Max != nil {
			max = strconv.FormatFloat(*element.Max, 'f', -1, 64)
		}
	case "string":
		if element.MinLength > 0 {
			min = strconv.FormatUint(element.MinLength, 10)
		}
		if element.MaxLength != nil {
			max = strconv.FormatUint(*element.MaxLength, 10)
		}
	}
	return min, max
}
func isOutOfBounds[V int | int32 | int64 | float64](value V, min, max string) (result bool) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int:
		minConverted, _ := strconv.Atoi(min)
		maxConverted, _ := strconv.Atoi(max)
		result = minConverted > any(value).(int) || maxConverted < any(value).(int)
	case reflect.Int32:
		minConverted, _ := strconv.ParseInt(min, 10, 32)
		maxConverted, _ := strconv.ParseInt(max, 10, 32)
		result = int32(minConverted) > any(value).(int32) || int32(maxConverted) < any(value).(int32)
	case reflect.Int64:
		minConverted, _ := strconv.ParseInt(min, 10, 64)
		maxConverted, _ := strconv.ParseInt(max, 10, 64)
		result = minConverted > any(value).(int64) || maxConverted < any(value).(int64)

	case reflect.Float64:
		minConverted, _ := strconv.ParseFloat(min, 64)
		maxConverted, _ := strconv.ParseFloat(max, 64)
		result = minConverted > any(value).(float64) || maxConverted < any(value).(float64)
	}
	return result
}
