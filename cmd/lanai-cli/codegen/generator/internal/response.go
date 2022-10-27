package internal

import (
	"github.com/getkin/kin-openapi/openapi3"
	"path"
)

type Responses openapi3.Responses

func (r Responses) ImportsUsed() (result []string) {
	var refs []string
	for _, response := range r {
		resp := Response(*response)
		// check if a struct will be created from this response
		if resp.CountFields() == 0 || (resp.CountFields() == 1 && resp.ContainsRef()) {
			break
		}
		if response.Ref != "" {
			refs = append(refs, path.Base(response.Ref))
		}
		for _, c := range response.Value.Content {
			if c.Schema.Ref != "" {
				refs = append(refs, path.Base(c.Schema.Ref))
			}
		}
	}
	return getImportsFromRef(refs)
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
