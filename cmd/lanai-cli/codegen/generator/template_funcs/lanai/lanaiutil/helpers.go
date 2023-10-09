package lanaiutil

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/util"
	"fmt"
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
	if err != nil && schema.Value.Type != openapi3.TypeString {
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
