package schematotext

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
	"github.com/getkin/kin-openapi/openapi3"
)

type translatorOptions struct {
	defaultObjectName     string
	currentPkg            string
	restrictExternalTypes bool
}

func WithDefaultObjectName(defaultObjectName string) func(options *translatorOptions) {
	return func(option *translatorOptions) {
		option.defaultObjectName = defaultObjectName
	}
}
func WithCurrentPackage(currentPkg string) func(options *translatorOptions) {
	return func(option *translatorOptions) {
		option.currentPkg = currentPkg
	}
}
func WithRestrictExternalTypes(restrictExternalTypes bool) func(options *translatorOptions) {
	return func(option *translatorOptions) {
		option.restrictExternalTypes = restrictExternalTypes
	}
}

type ToTextTranslator interface {
	ToText() (string, error)
}

func NewDataTypeToTextTranslator(element interface{}, opts ...func(option *translatorOptions)) ToTextTranslator {
	schema, _ := lanaiutil.ConvertToSchemaRef(element)
	if schema == nil {
		return NewDefaultType(element, opts...)
	}
	var translator ToTextTranslator
	switch schema.Value.Type {
	case openapi3.TypeNumber, openapi3.TypeInteger, openapi3.TypeBoolean:
		translator = NewBaseType(element, opts...)
	case openapi3.TypeString:
		translator = NewStringType(element, opts...)
	case openapi3.TypeArray:
		translator = NewArrayType(element, opts...)
	case openapi3.TypeObject:
		translator = NewObjectType(element, opts...)
	default:
		translator = NewDefaultType(element, opts...)
	}
	return translator
}
