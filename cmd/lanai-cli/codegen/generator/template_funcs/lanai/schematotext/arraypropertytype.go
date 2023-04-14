package schematotext

import "cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"

type arrayType struct {
	data                  interface{}
	defaultObjectName     string
	currentPkg            string
	restrictExternalTypes bool
}

func NewArrayType(data interface{}, opts ...func(option *translatorOptions)) *arrayType {
	o := &translatorOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return &arrayType{
		data:                  data,
		defaultObjectName:     o.defaultObjectName,
		currentPkg:            o.currentPkg,
		restrictExternalTypes: o.restrictExternalTypes,
	}
}

func (a arrayType) ToText() (string, error) {
	schema, err := lanaiutil.ConvertToSchemaRef(a.data)
	if err != nil {
		return "", err
	}
	inner, err := NewDataTypeToTextTranslator(
		schema.Value.Items,
		WithCurrentPackage(a.currentPkg),
		WithDefaultObjectName(a.defaultObjectName),
		WithRestrictExternalTypes(a.restrictExternalTypes)).ToText()
	if err != nil {
		return "", err
	}
	return "[]" + inner, nil
}
