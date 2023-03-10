package representation

import (
	"errors"
	"github.com/getkin/kin-openapi/openapi3"
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

func (o Operation) StructForMessage(messageType string, structRegistry map[string]string) (Struct, error) {
	switch strings.ToLower(messageType) {
	case "request":
		return o.requestStruct(structRegistry), nil
	case "response":
		return o.responseStruct(structRegistry), nil
	default:
		return Struct{}, errors.New("type must be \"request\" or \"response\"")
	}
}

type Struct struct {
	Import string
	Struct string
}

func (o Operation) requestStruct(structRegistry map[string]string) (result Struct) {
	defaultName := o.Name + "Request"
	defaultImport := structRegistry[strings.ToLower(defaultName)]
	if defaultImport != "" {
		result = Struct{
			Import: defaultImport,
			Struct: defaultName,
		}
		return result
	}

	paramRefs := Parameters(o.Data.Parameters).RefsUsed()
	if len(paramRefs) == 1 {
		result = Struct{
			Import: structRegistry[strings.ToLower(paramRefs[0])],
			Struct: paramRefs[0],
		}
		return result
	}

	if o.Data.RequestBody != nil {
		requestBodyRefs := RequestBody(*o.Data.RequestBody).RefsUsed()
		if len(requestBodyRefs) == 1 {
			result = Struct{
				Import: structRegistry[strings.ToLower(requestBodyRefs[0])],
				Struct: requestBodyRefs[0],
			}
			return result
		}
	}
	return result
}
func (o Operation) responseStruct(structRegistry map[string]string) (result Struct) {
	defaultName := o.Name + "Response"
	defaultImport := structRegistry[strings.ToLower(defaultName)]
	if defaultImport != "" {
		result = Struct{
			Import: defaultImport,
			Struct: defaultName,
		}
		return result
	}
	responses := Responses(o.Data.Responses).Sorted()
	responseRef := ""
	for _, response := range responses {
		refsUsed := response.RefsUsed()
		// Assume the last one is the "deepest" ref
		// eg. test.yaml: PostTestPath -> responses -> 200 -> GenericResponse
		if len(refsUsed) > 0 {
			responseRef = refsUsed[len(refsUsed)-1]
		}
		// Only interested in the first response
		break
	}
	if responseRef != "" {
		result = Struct{
			Import: structRegistry[strings.ToLower(responseRef)],
			Struct: responseRef,
		}
	}
	return result
}

func (o Operation) RefsUsed() (result []string) {
	for _, p := range o.Data.Parameters {
		result = append(result, p.Ref)
	}
	if o.Data.RequestBody != nil {
		r := RequestBody(*o.Data.RequestBody)
		result = append(result, r.RefsUsed()...)
	}

	return result
}
