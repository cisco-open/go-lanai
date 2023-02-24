package representation

import (
	"github.com/getkin/kin-openapi/openapi3"
	"path"
)

type Parameters openapi3.Parameters

func FromParameterMap(paramMap openapi3.ParametersMap) Parameters {
	result := Parameters{}
	for _, j := range paramMap {
		result = append(result, j)
	}
	return result
}
func (r Parameters) CountFields() int {
	return len(r)
}

func (r Parameters) ContainsRef() bool {
	for _, param := range r {
		if param.Ref != "" {
			return true
		}
	}
	return false
}

func (r Parameters) RefsUsed() []string {
	if r.CountFields() == 0 {
		return nil
	}
	var refs []string
	for _, schema := range r.schemas() {
		if schema.Ref != "" {
			refs = append(refs, path.Base(schema.Ref))
		}
	}
	return refs
}

func (r Parameters) ExternalImports() []string {
	if r.CountFields() == 0 {
		return nil
	}
	var imports []string
	for _, parameter := range r {
		if parameter.Ref == "" && parameter.Value.In != "path" {
			for _, schema := range parameterSchemas(parameter) {
				if isUUID(schema) {
					imports = append(imports, UUID_IMPORT_PATH)
				}
			}
		}
	}
	return imports
}

func (r Parameters) schemas() (result []*openapi3.SchemaRef) {
	for _, parameter := range r {
		result = append(result, parameterSchemas(parameter)...)
	}
	return result
}

func parameterSchemas(parameter *openapi3.ParameterRef) (result []*openapi3.SchemaRef) {
	result = append(result, parameter.Value.Schema)
	for _, a := range parameter.Value.Schema.Value.AllOf {
		result = append(result, a)
	}
	return result
}
