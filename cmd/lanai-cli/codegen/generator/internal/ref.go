package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/internal/representation"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
)

func importsUsedByPath(pathItem openapi3.PathItem, repositoryPath string) []string {
	var allImports []string
	for _, operation := range pathOperations(pathItem) {
		responses := representation.Responses(operation.Responses)
		parameters := representation.Parameters(operation.Parameters)
		var requestBody representation.RequestBody
		if operation.RequestBody != nil {
			requestBody = representation.RequestBody(*operation.RequestBody)
		}

		refs := responses.RefsUsed()
		numFieldsInRequestStruct := parameters.CountFields() + requestBody.CountFields()
		if numFieldsInRequestStruct != 1 {
			refs = append(refs, parameters.RefsUsed()...)
			refs = append(refs, requestBody.RefsUsed()...)
		}

		for _, i := range getImportsFromRef(refs) {
			allImports = append(allImports, repositoryPath+"/"+i)
		}

		// Grab any external dependencies
		allImports = append(allImports, responses.ExternalImports()...)
		allImports = append(allImports, parameters.ExternalImports()...)
		allImports = append(allImports, requestBody.ExternalImports()...)

	}

	uniqueImports := make(map[string]bool)
	for _, r := range allImports {
		uniqueImports[r] = true
	}

	var result []string
	for k := range uniqueImports {
		result = append(result, k)
	}
	return result
}

func getImportsFromRef(refs []string) []string {
	var imports []string
	for _, ref := range refs {
		loc := structLocation(ref)
		if loc != "" {
			imports = append(imports, loc)
		}
	}
	return imports
}

type RefChecker interface {
	CountFields() int
	ContainsRef() bool
}

func refCheckerFactory(element interface{}) (result []RefChecker, err error) {
	switch getInterfaceType(element) {
	case ResponseRefPtr:
		result = append(result, representation.Response(*element.(*openapi3.ResponseRef)))
	case OperationPtr:
		// Assume this is for Requests, so give requestbodies & parameters
		op := element.(*openapi3.Operation)
		var r representation.RequestBody
		if op.RequestBody != nil {
			r = representation.RequestBody(*op.RequestBody)
		}
		p := representation.Parameters(op.Parameters)
		result = append(result, r)
		result = append(result, p)
	default:
		return nil, fmt.Errorf("element not supported: %v", reflect.TypeOf(element))
	}

	return result, nil
}
func containsSingularRef(element interface{}) (bool, error) {
	fieldCount := 0
	containsRef := false
	r, err := refCheckerFactory(element)
	if err != nil {
		return false, err
	}
	for _, b := range r {
		fieldCount += b.CountFields()
		containsRef = containsRef || b.ContainsRef()
	}
	return fieldCount == 1 && containsRef, nil
}

func isEmpty(element interface{}) (bool, error) {
	count := 0
	r, err := refCheckerFactory(element)
	if err != nil {
		return false, err
	}
	for _, b := range r {
		count += b.CountFields()
	}

	return count == 0, nil
}
