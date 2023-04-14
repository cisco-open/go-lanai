package lanai

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/schematotext"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
)

func SchemaToText(element interface{}, defaultObjectName string, currentPkg string) (string, error) {
	dataType := schematotext.NewDataTypeToTextTranslator(
		element,
		schematotext.WithCurrentPackage(currentPkg),
		schematotext.WithDefaultObjectName(defaultObjectName),
		schematotext.WithRestrictExternalTypes(reflect.TypeOf(element) == reflect.TypeOf(&openapi3.Parameter{})))
	return dataType.ToText()
}
