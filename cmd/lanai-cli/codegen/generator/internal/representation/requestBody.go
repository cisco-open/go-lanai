package representation

import (
	"github.com/getkin/kin-openapi/openapi3"
	"path"
)

type RequestBody openapi3.RequestBodyRef

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

func (r RequestBody) RefsUsed() (result []string) {
	if r.CountFields() == 0 {
		return
	}
	if r.Ref != "" {
		result = append(result, path.Base(r.Ref))
	}

	if r.Value == nil {
		return
	}
	//Assumption - Responses will have just one mediatype defined in contract, e.g just "application/json"
	if len(r.Value.Content) > 1 {
		logger.Warn("more than one mediatype defined in requestBody, undefined behavior may occur")
	}
	for _, j := range r.Value.Content {
		if j.Schema.Ref != "" {
			result = append(result, path.Base(j.Schema.Ref))
		} else {
			for _, a := range j.Schema.Value.AllOf {
				if a.Ref != "" {
					result = append(result, path.Base(a.Ref))
				}
			}
		}
	}
	return result
}
