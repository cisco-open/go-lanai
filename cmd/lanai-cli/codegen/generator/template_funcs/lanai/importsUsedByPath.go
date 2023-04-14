package lanai

import (
	_go "cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/go"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
	"github.com/getkin/kin-openapi/openapi3"
)

func ImportsUsedByPath(pathItem openapi3.PathItem, repositoryPath string) []string {
	var allImports []string
	for _, operation := range pathItem.Operations() {
		responses := Responses{openapi.Responses(operation.Responses)}
		parameters := Parameters{openapi.Parameters(operation.Parameters)}
		var requestBody RequestBody
		if operation.RequestBody != nil {
			requestBody = RequestBody{openapi.RequestBody(*operation.RequestBody)}
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
		loc := _go.StructLocation(ref)
		if loc != "" {
			imports = append(imports, loc)
		}
	}
	return imports
}
