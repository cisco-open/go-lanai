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
		if schema.Ref == "" {
			result = append(result, schema.Value.Properties)
		} else {
			for _, a := range schema.Value.AllOf {
				if a.Ref == "" {
					result = append(result, a.Value.Properties)
				}
			}
		}
	}

	for _, requestBody := range c.RequestBodies {
		if requestBody.Value == nil {
			continue
		}
		for _, j := range requestBody.Value.Content {
			if j.Schema.Ref == "" {
				result = append(result, j.Schema.Value.Properties)
			} else {
				for _, a := range j.Schema.Value.AllOf {
					if a.Ref != "" {
						result = append(result, a.Value.Properties)
					}
				}
			}
		}
	}

	for _, parameter := range c.Parameters {
		if parameter.Ref == "" {
			result = append(result, parameter.Value.Schema.Value.Properties)
		} else {
			for _, a := range parameter.Value.Schema.Value.AllOf {
				if a.Ref == "" {
					result = append(result, a.Value.Properties)
				}
			}
		}
	}
	return result
}
