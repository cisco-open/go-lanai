package openapi

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/util"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
)

func requiredList(val interface{}) ([]string, error) {
	var list []string
	switch val.(type) {
	case *openapi3.SchemaRef:
		list = val.(*openapi3.SchemaRef).Value.Required
	case *openapi3.Parameter:
		parameter := val.(*openapi3.Parameter)
		if parameter.Required {
			list = append(list, parameter.Name)
		}
	case *openapi3.RequestBody:
		reqBody := val.(*openapi3.RequestBody)
		if reqBody.Required {
			list = append(list, "Body")
		}
	default:
		return nil, fmt.Errorf("requiredList error: unsupported interface %v", util.GetInterfaceType(val))
	}
	return list, nil
}
