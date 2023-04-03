package internal

import (
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

var (
	propertyFuncMap = template.FuncMap{
		"property":     NewProperty,
		"schemaToText": SchemaToText,
		"operation":    NewOperation,
		"schema":       NewSchema,
		"components":   NewComponents,
		"requestBody":  NewRequestBody,
	}
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
	return !listContains(p.RequiredList, p.PropertyName)
}
func PropertyToGoType(p Property) (string, error) {
	defaultObjectName := toTitle(p.TypePrefix + toTitle(p.PropertyName))
	result, err := SchemaToText(p.PropertyData, defaultObjectName, p.CurrentPkg)
	if err != nil {
		return "", err
	}
	return result, nil
}

func SchemaToText(element interface{}, defaultObjectName string, currentPkg string) (string, error) {
	dataType := getDataTypeToTextTranslator(
		element,
		WithCurrentPackage(currentPkg),
		WithDefaultObjectName(defaultObjectName),
		WithRestrictExternalTypes(reflect.TypeOf(element) == reflect.TypeOf(&openapi3.Parameter{})))
	return dataType.toText()
}

func ShouldHavePointer(p Property) bool {
	isRequired := listContains(p.RequiredList, p.PropertyName)
	schema, _ := convertToSchemaRef(p.PropertyData)
	if schema.Value.Enum != nil {
		return false
	}
	if len(schema.Value.Properties) == 0 {
		s := NewSchema("", schema)
		if s.HasAdditionalProperties() {
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

func valuePassesValidation(schema *openapi3.Schema, value reflect.Value) (result bool) {
	switch value.Kind() {
	case reflect.String:
		if rValue, _ := regex(*schema); rValue != nil {
			found, err := regexp.MatchString(rValue.Value, value.String())
			if err == nil || !found {
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
