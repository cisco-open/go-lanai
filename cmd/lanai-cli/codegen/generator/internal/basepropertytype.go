package internal

type BaseType struct {
	data interface{}
}

func NewBaseType(data interface{}, opts ...func(option *TranslatorOptions)) *BaseType {
	o := &TranslatorOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return &BaseType{
		data: data,
	}
}

func (b BaseType) toText() (string, error) {
	schema, err := convertToSchemaRef(b.data)
	if err != nil {
		return "", err
	}
	return schemaToGoBaseTypes(schema.Value).String(), nil
}
