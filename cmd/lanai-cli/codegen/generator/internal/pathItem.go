package internal

import "github.com/getkin/kin-openapi/openapi3"

func pathOperations(pathItem openapi3.PathItem) map[string]*openapi3.Operation {
	allOperations := map[string]*openapi3.Operation{
		"Get":     pathItem.Get,
		"Delete":  pathItem.Delete,
		"Post":    pathItem.Post,
		"Patch":   pathItem.Patch,
		"Connect": pathItem.Connect,
		"Head":    pathItem.Head,
		"Put":     pathItem.Put,
		"Options": pathItem.Options,
		"Trace":   pathItem.Trace,
	}

	ret := make(map[string]*openapi3.Operation)
	for k, v := range allOperations {
		if v != nil {
			ret[k] = v
		}
	}
	return ret
}
