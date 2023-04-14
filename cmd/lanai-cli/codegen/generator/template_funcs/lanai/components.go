package lanai

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
	"github.com/getkin/kin-openapi/openapi3"
)

type Components struct {
	openapi.Components
}

func NewComponents(data openapi3.Components) Components {
	return Components{openapi.Components(data)}
}

func (c Components) Imports() (result []string) {
	for _, schemas := range c.AllProperties() {
		//if any property uses a UUID, add the import in
		for _, properties := range schemas {
			result = append(result, lanaiutil.ExternalImportsFromFormat(properties)...)
			if len(properties.Value.Properties) != 0 {
				if SchemaRef(*properties).HasAdditionalProperties() {
					result = append(result, lanaiutil.JSON_IMPORT_PATH)
				}

			}
		}
	}
	return result
}
