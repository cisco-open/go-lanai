package internal

import (
	"github.com/getkin/kin-openapi/openapi3"
	"path"
)

type RequestBody openapi3.RequestBodyRef

func (r RequestBody) ImportsUsed() (result []string) {
	if r.CountFields() == 0 {
		return
	}
	var refs []string
	if r.Ref != "" {
		refs = append(refs, path.Base(r.Ref))
	}

	if r.Value == nil {
		return
	}
	for _, j := range r.Value.Content {
		if j.Schema.Ref != "" {
			refs = append(refs, path.Base(j.Schema.Ref))
		}
	}
	return getImportsFromRef(refs)
}
func (r RequestBody) ContainsRef() (result bool) {
	if r.Ref != "" {
		return true
	}
	if r.Value == nil {
		return false
	}
	for _, j := range r.Value.Content {
		if j.Schema.Ref != "" {
			result = true
		}
	}
	return result
}

func (r RequestBody) CountFields() (result int) {
	if r.Ref != "" {
		result++
	}
	if r.Value != nil {
		result += len(r.Value.Content)
	}
	return result
}
