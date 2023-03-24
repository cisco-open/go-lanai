package internal

import (
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

var (
	helperFuncMap = template.FuncMap{
		"args":         args,
		"increment":    increment,
		"listContains": listContains,
		"log":          templateLog,
		"derefBoolPtr": derefBoolPtr,
	}
)

func args(values ...interface{}) []interface{} {
	return values
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

func getInterfaceType(val interface{}) string {
	t := reflect.TypeOf(val)
	var res string
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
		res += "*"
	}
	return res + t.Name()
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

func increment(val int) int {
	return val + 1
}

func listContains(list []string, needle string) bool {
	for _, required := range list {
		if needle == required {
			return true
		}
	}
	return false
}

func templateLog(message ...interface{}) string {
	logger.Infof("%v", message)
	return ""
}

func derefBoolPtr(ptr *bool) (bool, error) {
	if ptr == nil {
		return false, errors.New("pointer is nil")
	}
	return *ptr, nil
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
