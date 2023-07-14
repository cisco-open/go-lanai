package lanai

//Representations of common structs in the templates, letting them look a little cleaner

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
	"github.com/getkin/kin-openapi/openapi3"
	"text/template"
)

var (
	propertyFuncMap = template.FuncMap{}
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

func (s Schema) AllSchemas() openapi3.SchemaRefs {
	return SchemaRef(*s.Data).AllSchemas()
}
func (s Schema) StructProperties() (result openapi3.SchemaRefs) {
	return SchemaRef(*s.Data).StructProperties()
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
	return openapi.SchemaRef(*s.Data).HasAdditionalProperties()
}
