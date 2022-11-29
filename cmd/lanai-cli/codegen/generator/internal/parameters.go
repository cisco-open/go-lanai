package internal

import (
	"github.com/getkin/kin-openapi/openapi3"
	"path"
)

type Parameters openapi3.Parameters

func (r Parameters) ImportsUsed() (result []string) {
	if r.CountFields() == 0 {
		return
	}
	var refs []string
	for _, param := range r {
		if param.Ref != "" {
			refs = append(refs, path.Base(param.Ref))
		}
	}
	return getImportsFromRef(refs)
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
