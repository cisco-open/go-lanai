package openapi

import "github.com/getkin/kin-openapi/openapi3"

type Components openapi3.Components

func (c Components) AllProperties() (result []openapi3.Schemas) {
	var schemas openapi3.SchemaRefs
	for _, schema := range c.Schemas {
		s := SchemaRef(*schema)
		schemas = append(schemas, s.AllSchemas()...)
	}

	for _, requestBody := range c.RequestBodies {
		if requestBody.Value == nil {
			continue
		}
		schemas = append(schemas, RequestBody(*requestBody).Schemas()...)
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
