package internal

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
		if parameter.Ref == "" && parameter.Value.In != "path" && parameter.Value.In != "query" {
			for _, schema := range _SchemaRef(*parameter.Value.Schema).AllSchemas() {
				if schema != nil {
					imports = append(imports, _SchemaRef(*schema).ExternalImports()...)
				}
			}
		}
	}
	return imports
}

func (r Parameters) schemas() (result openapi3.SchemaRefs) {
	for _, parameter := range r {
		result = append(result, _SchemaRef(*parameter.Value.Schema).AllSchemas()...)
	}
	return result
}
