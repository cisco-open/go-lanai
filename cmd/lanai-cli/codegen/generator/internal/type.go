package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/internal/representation"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func propertyType(property representation.Property) (string, error) {
	defaultObjectName := toTitle(property.TypePrefix + toTitle(property.PropertyName))
	return schemaToText(property.PropertyData, defaultObjectName)
}

func convertToSchemaRef(element interface{}) (*openapi3.SchemaRef, error) {
	var val *openapi3.SchemaRef
	interfaceType := getInterfaceType(element)
	switch interfaceType {
	case SchemaRefPtr:
		val = element.(*openapi3.SchemaRef)
	case ParameterPtr:
		val = element.(*openapi3.Parameter).Schema
	default:
		return nil, fmt.Errorf("convertToSchemaRef: unsupported interface %v", interfaceType)
	}
	return val, nil
}

func shouldBeUUIDType(element interface{}) bool {
	schema, err := convertToSchemaRef(element)
	if err != nil && schema.Value.Type != openapi3.TypeString {
		return false
	}

	formatIsUUID := strings.ToLower(schema.Value.Pattern) == "uuid" || strings.ToLower(schema.Value.Format) == "uuid"
	// exclude path parameters because go's validation only supports base types, so this should stay as a string
	isNotInPathParameter := getInterfaceType(element) != ParameterPtr || element.(*openapi3.Parameter).In != "path"
	return formatIsUUID && isNotInPathParameter
}

func schemaToText(element interface{}, defaultObjectName string) (result string, err error) {
	schema, err := convertToSchemaRef(element)
	if err != nil {
		return "", err
	}
	// for the non-object things
	switch schema.Value.Type {
	case openapi3.TypeBoolean:
		result = reflect.TypeOf(true).String()
	case openapi3.TypeNumber:
		result = reflect.TypeOf(1.1).String()
	case openapi3.TypeInteger:
		result = reflect.TypeOf(1).String()
	case openapi3.TypeString:
		if shouldBeUUIDType(element) {
			result = reflect.TypeOf(uuid.UUID{}).String()
		} else {
			result = reflect.TypeOf("string").String()
		}
	case openapi3.TypeArray:
		inner, err := schemaToText(schema.Value.Items, defaultObjectName)
		if err != nil {
			return "", err
		}
		result = "[]" + inner
	case openapi3.TypeObject:
		fallthrough
	default:
		if schema.Ref != "" {
			result = path.Base(schema.Ref)
		} else {
			result = defaultObjectName
		}
	}

	return result, nil
}

func schemaToGoType(val *openapi3.Schema) (result reflect.Type) {
	switch val.Type {
	case openapi3.TypeBoolean:
		result = reflect.TypeOf(true)
	case openapi3.TypeNumber:
		result = reflect.TypeOf(1.1)
	case openapi3.TypeInteger:
		result = reflect.TypeOf(1)
	case openapi3.TypeString:
		result = reflect.TypeOf("string")
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

func shouldHavePointer(element interface{}, isRequired bool) (bool, error) {
	schema, err := convertToSchemaRef(element)
	if err != nil {
		return false, err
	}
	if schema.Value.Type == "object" && isRequired {
		return true, nil
	}
	if schema.Value.Nullable {
		return true, nil
	}
	if schema.Value.Enum != nil {
		return false, nil
	}
	return valuePassesValidation(schema.Value, zeroValue(schema.Value)), nil
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
		fallthrough
	default:
		return valueIsWithinSchemaLimits(schema, value)
	}
}

func valueIsWithinSchemaLimits(schema *openapi3.Schema, value reflect.Value) bool {
	min, max := limitsForSchema(schema)
	switch value.Kind() {
	case reflect.String:
		return !isOutOfBounds(len(value.String()), min, max)
	case reflect.Int:
		return !isOutOfBounds(value.Interface().(int), min, max)
	case reflect.Float64:
		return !isOutOfBounds(value.Float(), min, max)
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		return !isOutOfBounds(value.Len(), min, max)
	}

	return false
}

func isOutOfBounds[V int | float64](value V, min, max string) (result bool) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int:
		minConverted, _ := strconv.Atoi(min)
		maxConverted, _ := strconv.Atoi(max)
		result = minConverted > any(value).(int) || maxConverted < any(value).(int)

	case reflect.Float64:
		minConverted, _ := strconv.ParseFloat(min, 64)
		maxConverted, _ := strconv.ParseFloat(max, 64)
		result = minConverted > any(value).(float64) || maxConverted < any(value).(float64)
	}
	return result
}
