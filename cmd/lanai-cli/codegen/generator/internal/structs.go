package internal

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"strings"
	"text/template"
)

var (
	structsFuncMap = template.FuncMap{
		"propertyType":        propertyType,
		"structTag":           structTag,
		"requiredList":        requiredList,
		"containsSingularRef": containsSingularRef,
		"shouldHavePointer":   shouldHavePointer,
		"defaultNameFromPath": defaultNameFromPath,
		"registerStruct":      registerStruct,
		"structLocation":      structLocation,
		"structRegistry":      getStructRegistry,
		"importsUsedByPath":   importsUsedByPath,
		"isEmpty":             isEmpty,
		"pathOperations":      pathOperations,
	}
)

func requiredList(val interface{}) ([]string, error) {
	var list []string
	interfaceType := getInterfaceType(val)
	switch interfaceType {
	case SchemaRefPtr:
		list = val.(*openapi3.SchemaRef).Value.Required
	case ParameterPtr:
		parameter := val.(*openapi3.Parameter)
		if parameter.Required {
			list = append(list, parameter.Name)
		}
	default:
		return nil, fmt.Errorf("requiredList error: unsupported interface %v", interfaceType)
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
	return structRegistry[strings.ToLower(schemaName)]
}

func getStructRegistry() map[string]string {
	return structRegistry
}
