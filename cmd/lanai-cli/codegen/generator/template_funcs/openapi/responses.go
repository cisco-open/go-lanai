package openapi

import (
	"github.com/getkin/kin-openapi/openapi3"
	"sort"
)

type Responses openapi3.Responses

func (r Responses) RefsUsed() (result []string) {
	for _, response := range r.Sorted() {
		// check if a struct will be created from this response
		if response.CountFields() == 0 || (response.CountFields() == 1 && response.ContainsRef()) {
			break
		}
		result = append(result, response.RefsUsed()...)
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
