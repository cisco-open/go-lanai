package openapi

import "github.com/getkin/kin-openapi/openapi3"

type SchemaRef openapi3.SchemaRef

// AllSchemas will return every schemaRef object associated with this object
// Will Include:
// - The base schema
// - All the schemas part of the AllOf array (recursively)
// - Items, if the schema is an array (recursively)
func (s SchemaRef) AllSchemas() (result openapi3.SchemaRefs) {
	base := openapi3.SchemaRef(s)
	result = append(result, &base)
	for _, a := range s.Value.AllOf {
		result = append(result, SchemaRef(*a).AllSchemas()...)
	}
	if s.Value.Type == "array" {
		result = append(result, SchemaRef(*s.Value.Items).AllSchemas()...)
	}

	return
}

// HasAdditionalProperties returns true if the schema supports additionalProperties of any kind
func (s SchemaRef) HasAdditionalProperties() bool {
	additionalPropertiesAllowed := s.Value.AdditionalPropertiesAllowed != nil && *s.Value.AdditionalPropertiesAllowed
	additionalPropsDefined := s.Value.AdditionalProperties != nil

	return additionalPropertiesAllowed || additionalPropsDefined
}
