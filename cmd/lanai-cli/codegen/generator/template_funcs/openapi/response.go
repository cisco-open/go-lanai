package openapi

import (
	"github.com/getkin/kin-openapi/openapi3"
	"path"
)

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
		// Double check that this isn't an empty wrapper around another ref
		if r.CountFields() != 1 {
			refs = append(refs, path.Base(r.Ref))
		}
	}

	for _, schema := range r.Schemas() {
		if schema.Ref != "" {
			refs = append(refs, path.Base(schema.Ref))
		}
	}

	return refs
}

func (r Response) Schemas() (result []*openapi3.SchemaRef) {
	for _, c := range r.Value.Content {
		result = append(result, SchemaRef(*c.Schema).AllSchemas()...)
	}
	return result
}
