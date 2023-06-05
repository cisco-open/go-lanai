package internal

type ArrayType struct {
	data                  interface{}
	defaultObjectName     string
	currentPkg            string
	restrictExternalTypes bool
}

func NewArrayType(data interface{}, opts ...func(option *TranslatorOptions)) *ArrayType {
	o := &TranslatorOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return &ArrayType{
		data:                  data,
		defaultObjectName:     o.defaultObjectName,
		currentPkg:            o.currentPkg,
		restrictExternalTypes: o.restrictExternalTypes,
	}
}

func (a ArrayType) toText() (string, error) {
	schema, err := convertToSchemaRef(a.data)
	if err != nil {
		return "", err
	}
	inner, err := getDataTypeToTextTranslator(
		schema.Value.Items,
		WithCurrentPackage(a.currentPkg),
		WithDefaultObjectName(a.defaultObjectName),
		WithRestrictExternalTypes(a.restrictExternalTypes)).toText()
	if err != nil {
		return "", err
	}
	return "[]" + inner, nil
}
