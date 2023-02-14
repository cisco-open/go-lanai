package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/internal/representation"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"strings"
)

func structTag(property representation.Property, requiredParams []string) string {
	nameType := nameType(property.PropertyData)
	result := fmt.Sprintf("`%v:\"%v\"", nameType, property.PropertyName)
	schema, err := convertToSchemaRef(property.PropertyData)
	if err != nil {
		return ""
	}
	binding := bindings(property.PropertyName, schema, requiredParams)
	if binding != "" {
		result = fmt.Sprintf("%v binding:\"%v\"", result, binding)
	}
	return result
}

func nameType(element interface{}) string {
	nameType := "json"

	interfaceType := getInterfaceType(element)
	switch interfaceType {
	case SchemaRefPtr:
		nameType = "json"
	case ParameterPtr:
		switch element.(*openapi3.Parameter).In {
		case "query":
			nameType = "form"
		case "path":
			fallthrough
		default:
			nameType = "uri"
		}
	default:
		logger.Errorf("no supported added for struct-tags for %v, double check contract or log a bug", interfaceType)
	}
	return nameType
}

func bindings(propertyName string, element *openapi3.SchemaRef, requiredList []string) string {
	var bindingParts []string

	bindingParts = append(bindingParts, omitEmptyTags(propertyName, requiredList, element)...)
	bindingParts = append(bindingParts, requiredTag(propertyName, requiredList)...)
	bindingParts = append(bindingParts, validationTags(element, requiredList)...)
	return strings.Join(bindingParts, ",")
}

func validationTags(element *openapi3.SchemaRef, requiredList []string) []string {
	result := regexTag(element)
	result = append(result, limitTags(element)...)
	result = append(result, enumOf(element.Value.Enum)...)
	if element.Value.Type == "array" {
		innerParts := validationTags(element.Value.Items, requiredList)
		if innerParts != nil {
			result = append(result, "dive")
			result = append(result, innerParts...)
		}
	}
	return result
}

func requiredTag(propertyName string, requiredList []string) (result []string) {
	if listContains(requiredList, propertyName) {
		result = append(result, "required")
	}
	return result
}

func regexTag(element *openapi3.SchemaRef) (result []string) {
	if element == nil {
		return result
	}
	rValue := regex(*element.Value)
	if rValue != nil && rValue.Value != "" {
		result = append(result, generateNameFromRegex(rValue.Value))
	}
	return result
}

func limitTags(element *openapi3.SchemaRef) (result []string) {
	if element == nil {
		return result
	}

	min, max := limitsForSchema(element.Value)
	if min != "" {
		result = append(result, fmt.Sprintf("min=%v", min))
	}
	if max != "" {
		result = append(result, fmt.Sprintf("max=%v", max))
	}
	return result
}

func omitEmptyTags(propertyName string, requiredList []string, schemaRef *openapi3.SchemaRef) (result []string) {
	if schemaRef == nil {
		return result
	}

	if !listContains(requiredList, propertyName) && valuePassesValidation(schemaRef.Value, zeroValue(schemaRef.Value)) {
		result = append(result, "omitempty")
	}
	return result
}

func enumOf(enums []interface{}) (result []string) {
	allEnums := ""
	for _, e := range enums {
		if e == nil {
			continue
		}
		allEnums += fmt.Sprintf("%v ", e.(string))
	}

	if allEnums == "" {
		return result
	}
	binding := strings.TrimSpace("enumof=" + allEnums)
	result = append(result, binding)
	return result
}
