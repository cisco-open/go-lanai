package representation

import (
	"github.com/getkin/kin-openapi/openapi3"
	"path"
	"sort"
)

type Responses openapi3.Responses

func (r Responses) RefsUsed() (result []string) {
	for _, response := range r {
		resp := Response(*response)
		// check if a struct will be created from this response
		if resp.CountFields() == 0 || (resp.CountFields() == 1 && resp.ContainsRef()) {
			break
		}
		result = append(result, resp.RefsUsed()...)
	}
	return result
}

func (r Responses) Sorted() (result []Response) {
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		result = append(result, Response(*r[k]))
	}

	return result
}

type Response openapi3.ResponseRef

func (r Response) CountFields() (result int) {
	content := Content(r.Value.Content)
	result += content.CountFields()
	return result
}

func (r Response) ContainsRef() bool {
	result := r.Ref != ""
	if !result {
		content := Content(r.Value.Content)
		result = content.ContainsRef()
	}
	return result
}

func (r Response) RefsUsed() (result []string) {
	var refs []string
	if r.Ref != "" {
		refs = append(refs, path.Base(r.Ref))
	}
	for _, c := range r.Value.Content {
		if c.Schema.Ref != "" {
			refs = append(refs, path.Base(c.Schema.Ref))
		}
	}

	return refs
}
