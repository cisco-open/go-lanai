package internal

import (
	"github.com/getkin/kin-openapi/openapi3"
)

type Components openapi3.Components

func NewComponents(data openapi3.Components) Components {
	return Components(data)
}

func (c Components) Imports() (result []string) {
	for _, schemas := range c.AllProperties() {
		//if any property uses a UUID, add the import in
		for _, properties := range schemas {
			result = append(result, externalImportsFromFormat(properties)...)
			if len(properties.Value.Properties) != 0 {
				if _SchemaRef(*properties).HasAdditionalProperties() {
					result = append(result, JSON_IMPORT_PATH)
				}

			}
		}
	}
	return result
}

func (c Components) AllProperties() (result []openapi3.Schemas) {
	var schemas openapi3.SchemaRefs
	for _, schema := range c.Schemas {
		s := Schema{Data: schema}
		schemas = append(schemas, s.AllSchemas()...)
	}

	for _, requestBody := range c.RequestBodies {
		if requestBody.Value == nil {
			continue
		}
		schemas = append(schemas, RequestBody(*requestBody).schemas()...)
	}

	schemas = append(schemas, FromParameterMap(c.Parameters).schemas()...)
	for _, s := range schemas {
		// If there is a Ref, that doesn't count as a property
		if s.Ref == "" {
			result = append(result, s.Value.Properties)
		}
	}
	return
}
