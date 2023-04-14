package lanai

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
)

type Parameters struct {
	openapi.Parameters
}

func (r Parameters) ExternalImports() []string {
	if r.CountFields() == 0 {
		return nil
	}
	var imports []string
	for _, parameter := range r.Parameters {
		if parameter.Ref == "" && parameter.Value.In != "path" && parameter.Value.In != "query" {
			for _, schema := range SchemaRef(*parameter.Value.Schema).AllSchemas() {
				imports = append(imports, lanaiutil.ExternalImportsFromFormat(schema)...)
			}
		}
	}
	return imports
}
