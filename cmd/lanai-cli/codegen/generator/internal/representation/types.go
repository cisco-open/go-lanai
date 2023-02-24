package representation

//Representations of common structs in the templates, letting them look a little cleaner

import (
	"github.com/getkin/kin-openapi/openapi3"
)

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

func (s Schema) AllSchemaRefs() (result openapi3.SchemaRefs) {
	if s.Data.Value.AllOf == nil || s.Data.Ref != "" {
		result = append(result, s.Data)
	} else {
		for _, schemaRef := range s.Data.Value.AllOf {
			schema := NewSchema(s.Name, schemaRef)
			result = append(result, schema.AllSchemaRefs()...)
		}
	}

	return result
}
