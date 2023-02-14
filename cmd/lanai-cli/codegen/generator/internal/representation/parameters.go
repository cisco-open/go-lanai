package representation

import (
	"github.com/getkin/kin-openapi/openapi3"
	"path"
)

type Parameters openapi3.Parameters

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
	for _, param := range r {
		if param.Ref != "" {
			refs = append(refs, path.Base(param.Ref))
		}
	}
	return refs
}
