package representation

import (
	"github.com/getkin/kin-openapi/openapi3"
)

type Components openapi3.Components

func NewComponents(data openapi3.Components) Components {
	return Components(data)
}

func (c Components) Imports() (result []string) {
	for _, schemas := range c.allProperties() {
		//if any property uses a UUID, add the import in
		for _, properties := range schemas {
			if isUUID(properties) {
				result = append(result, UUID_IMPORT_PATH)
				return
			}
		}
	}
	return result
}

func (c Components) allProperties() (result []openapi3.Schemas) {
	for _, schema := range c.Schemas {
		s := Schema{Data: schema}
		for _, j := range s.AllSchemaRefs() {
			if j.Ref == "" {
				result = append(result, j.Value.Properties)
			}
		}
	}

	for _, requestBody := range c.RequestBodies {
		if requestBody == nil || requestBody.Value == nil {
			continue
		}
		r := RequestBody(*requestBody)
		for _, j := range r.schemas() {
			if j.Ref == "" {
				result = append(result, j.Value.Properties)
			}
		}
	}

	parameters := FromParameterMap(c.Parameters)
	for _, parameter := range parameters.schemas() {
		if parameter.Ref == "" {
			result = append(result, parameter.Value.Properties)
		}
	}
	return result
}
