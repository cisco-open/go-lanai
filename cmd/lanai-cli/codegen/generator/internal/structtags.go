package internal

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"strings"
)

func structTags(p Property) string {
	requiredParams := p.RequiredList
	nameType := nameType(p.PropertyData)
	name := p.PropertyName
	if p.OmitJSON {
		name = "-"
	}
	result := fmt.Sprintf("%v:\"%v\"", nameType, name)
	binding := bindings(p.PropertyName, p.PropertyData, requiredParams)
	if binding != "" {
		result = fmt.Sprintf("%v binding:\"%v\"", result, binding)
	}

	result = fmt.Sprintf("`%v`", result)
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

func bindings(propertyName string, data interface{}, requiredList []string) string {
	var bindingParts []string
	validationTags := validationTags(data, requiredList)
	bindingParts = append(bindingParts, omitEmptyTags(propertyName, requiredList, len(validationTags))...)
	bindingParts = append(bindingParts, requiredTag(propertyName, requiredList)...)
	bindingParts = append(bindingParts, validationTags...)
	return strings.Join(bindingParts, ",")
}

func validationTags(data interface{}, requiredList []string) []string {
	result := regexTag(data)
	schemaRef, err := convertToSchemaRef(data)
	if err != nil {
		return nil
	}
	result = append(result, limitTags(schemaRef)...)
	result = append(result, enumOf(schemaRef.Value.Enum)...)
	if schemaRef.Value.Type == "array" {
		innerParts := validationTags(schemaRef.Value.Items, requiredList)
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

func regexTag(element interface{}) (result []string) {
	if element == nil || shouldBeUUIDType(element) {
		return result
	}
	schemaRef, err := convertToSchemaRef(element)
	if err != nil {
		return nil
	}
	rValue, _ := regex(*schemaRef.Value)
	if rValue != nil && rValue.Value != "" {
		result = append(result, generateNameFromRegex(rValue.Value))
	}
	return result
}

func limitTags(schemaRef *openapi3.SchemaRef) (result []string) {
	if schemaRef == nil {
		return result
	}

	min, max := limitsForSchema(schemaRef.Value)
	if min != "" {
		result = append(result, fmt.Sprintf("min=%v", min))
	}
	if max != "" {
		result = append(result, fmt.Sprintf("max=%v", max))
	}
	return result
}

// omitEmptyTags will adds omitEmpty tag if:
// the property is not the list of required properties
// numberOfValidationTags > 0 - if there are any validations that need to be omitted
func omitEmptyTags(propertyName string, requiredList []string, numberOfValidationTags int) (result []string) {
	if !listContains(requiredList, propertyName) && numberOfValidationTags > 0 {
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
