package internal

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"path"
	"strings"
	"text/template"
)

var (
	structsFuncMap = template.FuncMap{
		"requiredList":        requiredList,
		"containsSingularRef": containsSingularRef,
		"defaultNameFromPath": defaultNameFromPath,
		"registerStruct":      registerStruct,
		"structLocation":      structLocation,
		"structRegistry":      getStructRegistry,
		"importsUsedByPath":   importsUsedByPath,
		"isEmpty":             isEmpty,
		"pathOperations":      pathOperations,
		"structTags":          structTags,
		"shouldHavePointer":   ShouldHavePointer,
		"propertyToGoType":    PropertyToGoType,
	}
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
	default:
		return nil, fmt.Errorf("requiredList error: unsupported interface %v", getInterfaceType(val))
	}
	return list, nil
}

func defaultNameFromPath(val string) string {
	path := pathPart(val, PathAfterVersion)
	path = replaceParameterDelimiters(path, "/", "")
	pathParts := strings.Split(path, "/")

	// make this camelCase
	for p := range pathParts {
		if p == 0 {
			continue
		}
		pathParts[p] = toTitle(pathParts[p])
	}

	return strings.Join(pathParts, "")
}

var structRegistry = make(map[string]string)

func registerStruct(schemaName string, packageName string) string {
	structRegistry[strings.ToLower(schemaName)] = packageName
	return ""
}

func structLocation(schemaName string) string {
	return structRegistry[strings.ToLower(path.Base(schemaName))]
}

func getStructRegistry() map[string]string {
	return structRegistry
}
