package openapi

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

func (r Parameters) schemas() (result openapi3.SchemaRefs) {
	for _, parameter := range r {
		result = append(result, SchemaRef(*parameter.Value.Schema).AllSchemas()...)
	}
	return result
}
