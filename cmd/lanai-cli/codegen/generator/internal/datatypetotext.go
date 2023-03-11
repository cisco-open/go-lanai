package internal

import (
	"github.com/getkin/kin-openapi/openapi3"
)

type TranslatorOptions struct {
	defaultObjectName     string
	currentPkg            string
	restrictExternalTypes bool
}

func WithDefaultObjectName(defaultObjectName string) func(options *TranslatorOptions) {
	return func(option *TranslatorOptions) {
		option.defaultObjectName = defaultObjectName
	}
}
func WithCurrentPackage(currentPkg string) func(options *TranslatorOptions) {
	return func(option *TranslatorOptions) {
		option.currentPkg = currentPkg
	}
}
func WithRestrictExternalTypes(restrictExternalTypes bool) func(options *TranslatorOptions) {
	return func(option *TranslatorOptions) {
		option.restrictExternalTypes = restrictExternalTypes
	}
}

type ToTextTranslator interface {
	toText() (string, error)
}

func getDataTypeToTextTranslator(element interface{}, opts ...func(option *TranslatorOptions)) ToTextTranslator {
	schema, _ := convertToSchemaRef(element)
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
