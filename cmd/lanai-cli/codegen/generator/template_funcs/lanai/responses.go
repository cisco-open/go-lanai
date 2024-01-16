package lanai

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
	"github.com/getkin/kin-openapi/openapi3"
)

type Responses struct {
	openapi.Responses
}

func NewResponses(src *openapi3.Responses) *Responses {
	if src == nil {
		return &Responses{}
	}
	return &Responses{Responses: openapi.Responses{Responses: *src}}
}

func (r Responses) ExternalImports() (result []string) {
	for _, response := range r.Responses.Map() {
		resp := Response{openapi.Response(*response)}
		// check if a struct will be created from this response
		if resp.CountFields() == 0 || (resp.CountFields() == 1 && resp.ContainsRef()) {
			break
		}
		result = append(result, resp.ExternalImports()...)
	}
	return result
}

type Response struct {
	openapi.Response
}

func (r Response) ExternalImports() (result []string) {
	if r.Ref != "" {
		return result
	}
	for _, schema := range r.Schemas() {
		result = append(result, lanaiutil.ExternalImportsFromFormat(schema)...)
	}
	return result
}
