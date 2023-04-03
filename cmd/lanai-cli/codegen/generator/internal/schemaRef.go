package internal

import "github.com/getkin/kin-openapi/openapi3"

type _SchemaRef openapi3.SchemaRef

// AllSchemas will return every schemaRef object associated with this object
// Will Include:
// - The base schema
// - All the schemas part of the AllOf array (recursively)
// - Items, if the schema is an array (recursively)
func (s _SchemaRef) AllSchemas() (result openapi3.SchemaRefs) {
	base := openapi3.SchemaRef(s)
	result = append(result, &base)
	for _, a := range s.Value.AllOf {
		result = append(result, _SchemaRef(*a).AllSchemas()...)
	}
	if s.Value.Type == "array" {
		result = append(result, _SchemaRef(*s.Value.Items).AllSchemas()...)
	}

	return
}

// StructProperties will return the schemas to be turned into properties for the struct
// Includes:
// The base schema if:
// - it is not AllOf
// - or it is a ref
// - or has additionalProperties
// Individual schemas part of an AllOf that match the above, if the schema isn't already a Ref
func (s _SchemaRef) StructProperties() (result openapi3.SchemaRefs) {
	if s.Value.AllOf == nil || s.Ref != "" || s.HasAdditionalProperties() {
		base := openapi3.SchemaRef(s)
		result = append(result, &base)
	}

	if s.Value.AllOf != nil && s.Ref == "" {
		for _, schemaRef := range s.Value.AllOf {
			result = append(result, _SchemaRef(*schemaRef).StructProperties()...)
		}
	}

	return
}

// HasAdditionalProperties returns true if the schema supports additionalProperties of any kind
func (s _SchemaRef) HasAdditionalProperties() bool {
	additionalPropertiesAllowed := s.Value.AdditionalPropertiesAllowed != nil && *s.Value.AdditionalPropertiesAllowed
	additionalPropsDefined := s.Value.AdditionalProperties != nil

	return additionalPropertiesAllowed || additionalPropsDefined
}
