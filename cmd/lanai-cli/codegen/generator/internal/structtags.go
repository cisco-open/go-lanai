package internal

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"strings"
)

func structTags(p Property) string {
	result := ""
	switch p.PropertyData.(type) {
	case *openapi3.Parameter:
		if p.PropertyData.(*openapi3.Parameter).In == "header" {
			result = fmt.Sprintf("header:\"%v\" ", p.PropertyName)
		}
	}

	nameType := nameType(p.PropertyData)
	name := p.PropertyName
	if p.OmitJSON {
		name = "-"
	} else if nameType == "json" {
		schema, _ := convertToSchemaRef(p.PropertyData)
		if p.IsOptional() && schema.Value.Type != "array" && (ShouldHavePointer(p) || !zeroValueIsValid(schema)) {
			name = fmt.Sprintf("%v,omitempty", name)
		}
	}
	result = result + fmt.Sprintf("%v:\"%v\"", nameType, name)
	binding := bindings(p.PropertyName, p.PropertyData, p.RequiredList)
	if binding != "" {
		result = fmt.Sprintf("%v binding:\"%v\"", result, binding)
	}

	result = fmt.Sprintf("`%v`", result)
	return result
}

func nameType(element interface{}) string {
	nameType := "json"

	switch element.(type) {
	case *openapi3.SchemaRef:
		nameType = "json"
	case *openapi3.Parameter:
		switch element.(*openapi3.Parameter).In {
		case "query", "header":
			nameType = "form"
		case "path":
			fallthrough
		default:
			nameType = "uri"
		}
	default:
		logger.Errorf("no supported added for struct-tags for %v, double check contract or log a bug", getInterfaceType(element))
	}
	return nameType
}

func bindings(propertyName string, data interface{}, requiredList []string) string {
	bindingParts := requiredTag(propertyName, requiredList)
	if isBaseType(data) {
		validationTags := validationTags(data)
		bindingParts = append(bindingParts, omitEmptyTags(propertyName, requiredList, len(validationTags))...)
		bindingParts = append(bindingParts, validationTags...)
	}
	return strings.Join(bindingParts, ",")
}

func validationTags(data interface{}) []string {
	result := regexTag(data)
	schemaRef, err := convertToSchemaRef(data)
	if err != nil {
		return nil
	}
	result = append(result, limitTags(schemaRef)...)
	result = append(result, enumOf(schemaRef.Value.Enum)...)
	if schemaRef.Value.Type == openapi3.TypeArray {
		innerParts := validationTags(schemaRef.Value.Items)
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
	schemaRef, err := convertToSchemaRef(element)
	if err != nil {
		return nil
	}
	rValue, _ := regex(*schemaRef.Value)
	if rValue != nil {
		result = append(result, rValue.Name)
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
// the property is not in the list of required properties
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
