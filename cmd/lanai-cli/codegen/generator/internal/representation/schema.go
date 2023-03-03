package representation

//Representations of common structs in the templates, letting them look a little cleaner

import (
	"github.com/getkin/kin-openapi/openapi3"
)

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
	if s.Data.Value.AllOf == nil || s.Data.Ref != "" || s.HasAdditionalProperties() {
		result = append(result, s.Data)
	}
	if s.Data.Ref == "" {
		for _, schemaRef := range s.Data.Value.AllOf {
			schema := NewSchema(s.Name, schemaRef)
			result = append(result, schema.AllSchemaRefs()...)
		}
	}

	return result
}

func (s Schema) AllProperties() (result openapi3.Schemas) {
	result = make(openapi3.Schemas)
	if s.Data.Value.Type == openapi3.TypeObject || s.Data.Value.Type == "" {
		result = s.Data.Value.Properties
	}
	return result
}

func (s Schema) Type() string {
	if s.Data.Value.Type == "" {
		return openapi3.TypeObject
	}
	return s.Data.Value.Type
}

func (s Schema) HasAdditionalProperties() bool {
	additionalPropertiesAllowed := s.Data.Value.AdditionalPropertiesAllowed != nil && *s.Data.Value.AdditionalPropertiesAllowed
	additionalPropsDefined := s.Data.Value.AdditionalProperties != nil

	return additionalPropertiesAllowed || additionalPropsDefined
}
