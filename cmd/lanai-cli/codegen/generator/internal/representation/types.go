package representation

//Representations of common structs in the templates, letting them look a little cleaner

import "github.com/getkin/kin-openapi/openapi3"

type Property struct {
	PropertyName string
	TypePrefix   string
	PropertyData interface{}
}

func NewProperty(name string, data interface{}) Property {
	return Property{
		PropertyName: name,
		PropertyData: data,
	}
}

func PropertyTypePrefix(prop Property, prefix string) Property {
	prop.TypePrefix = prefix
	return prop
}

type Operation struct {
	Name string
	Data *openapi3.Operation
}

func NewOperation(defaultName string, data *openapi3.Operation) Operation {
	name := defaultName
	if data.OperationID != "" {
		name = data.OperationID
	}
	return Operation{
		Name: name,
		Data: data,
	}
}

type Schema struct {
	Name string
	Data *openapi3.SchemaRef
}

func NewSchema(name string, data *openapi3.SchemaRef) Schema {
	return Schema{
		Name: name,
		Data: data,
	}
}
