package internal

import (
	"errors"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
	"strconv"
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

func templateLog(message string) string {
	logger.Infof(message)
	return ""
}

func derefBoolPtr(ptr *bool) (bool, error) {
	if ptr == nil {
		return false, errors.New("pointer is nil")
	}
	return *ptr, nil
}
