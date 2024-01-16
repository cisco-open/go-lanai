package lanai

import (
	_go "cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/go"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
	"errors"
	"github.com/getkin/kin-openapi/openapi3"
	"path"
	"strings"
)

type Operation struct {
	Name string
	Data *openapi3.Operation
}

func NewOperation(data *openapi3.Operation, defaultName string) Operation {
	name := defaultName
	if data.OperationID != "" {
		name = data.OperationID
	}
	return Operation{
		Name: name,
		Data: data,
	}
}

func (o Operation) StructForMessage(messageType string, structRegistry map[string]string) (*_go.Struct, error) {
	switch strings.ToLower(messageType) {
	case "request":
		return o.RequestStruct(structRegistry), nil
	case "response":
		return o.ResponseStruct(structRegistry), nil
	default:
		return nil, errors.New("type must be \"request\" or \"response\"")
	}
}

func (o Operation) RequestRefsUsed() (result []string) {
	for _, p := range o.Data.Parameters {
		result = append(result, p.Ref)
	}
	if o.Data.RequestBody != nil {
		r := openapi.RequestBody(*o.Data.RequestBody)
		result = append(result, r.RefsUsed()...)
	}

	return result
}

func (o Operation) ResponseRefsUsed() (result []string) {
	responses := NewResponses(o.Data.Responses).Sorted()
	for _, resp := range responses {
		if resp.CountFields() == 1 && resp.ContainsRef() {
			result = append(result, resp.RefsUsed()...)
			break
		}
		break
	}
	return result
}

func (o Operation) RequestStruct(structRegistry map[string]string) *_go.Struct {
	structName := o.Name + "Request"
	var structPackage string
	p, ok := structRegistry[strings.ToLower(structName)]
	if ok {
		structPackage = p
	} else {
		refs := o.RequestRefsUsed()
		if refs == nil {
			return nil
		}
		singularRef := refs[0]
		structName = path.Base(singularRef)

		p, ok := structRegistry[strings.ToLower(structName)]
		if ok {
			structPackage = p
		}
	}
	return &_go.Struct{
		Package: structPackage,
		Name:    structName,
	}
}

func (o Operation) ResponseStruct(structRegistry map[string]string) *_go.Struct {
	structName := o.Name + "Response"
	structPackage := structRegistry[strings.ToLower(structName)]
	if structPackage == "" {
		refsUsed := o.ResponseRefsUsed()
		if len(refsUsed) == 0 {
			return nil
		}
		responseRef := refsUsed[0]
		structName = path.Base(responseRef)
		p, ok := structRegistry[strings.ToLower(structName)]
		if ok {
			structPackage = p
		}
	}
	return &_go.Struct{
		Package: structPackage,
		Name:    structName,
	}
}

func (o Operation) AllResponseContent() (result []*openapi3.MediaType) {
	responses := NewResponses(o.Data.Responses).Sorted()
	for _, response := range responses {
		for _, content := range response.Value.Content {
			result = append(result, content)
		}
	}
	return result
}
