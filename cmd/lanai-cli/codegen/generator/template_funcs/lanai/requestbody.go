package lanai

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
	"github.com/getkin/kin-openapi/openapi3"
)

type RequestBody struct {
	openapi.RequestBody
}

func NewRequestBody(body openapi3.RequestBodyRef) RequestBody {
	return RequestBody{openapi.RequestBody(body)}
}

func (r RequestBody) ExternalImports() (result []string) {
	if r.CountFields() == 0 {
		return
	}
	if r.Value == nil {
		return
	}
	for _, schema := range r.Schemas() {
		result = append(result, lanaiutil.ExternalImportsFromFormat(schema)...)
		if schema.Ref == "" {
			if schema.Value.Type != "" && schema.Value.Type != "object" {
				result = append(result, lanaiutil.JSON_IMPORT_PATH)
			}
		}
	}
	return result
}
