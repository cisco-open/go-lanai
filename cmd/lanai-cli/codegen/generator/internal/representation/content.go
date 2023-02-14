package representation

import "github.com/getkin/kin-openapi/openapi3"

type Content openapi3.Content

func (c *Content) ContainsRef() bool {
	for _, c := range *c {
		if c.Schema.Ref != "" {
			return true
		}
	}
	return false
}

func (c *Content) CountFields() (result int) {
	for _, content := range *c {
		if content.Schema.Ref != "" {
			result++
		} else {
			if content.Schema.Value.Type == "object" {
				result++
			}
			result += len(content.Schema.Value.AllOf)
		}
	}
	return result
}
