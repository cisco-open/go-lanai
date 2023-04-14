package lanai

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
	"github.com/getkin/kin-openapi/openapi3"
)

type SchemaRef openapi.SchemaRef

// StructProperties will return the schemas to be turned into properties for the struct
// Includes:
// The base schema if:
// - it is not AllOf
// - or it is a ref
// - or has additionalProperties
// Individual schemas part of an AllOf that match the above, if the schema isn't already a Ref
func (s SchemaRef) StructProperties() (result openapi3.SchemaRefs) {
	if s.Value.AllOf == nil || s.Ref != "" || s.HasAdditionalProperties() {
		base := openapi3.SchemaRef(s)
		result = append(result, &base)
	}

	if s.Value.AllOf != nil && s.Ref == "" {
		for _, schemaRef := range s.Value.AllOf {
			result = append(result, SchemaRef(*schemaRef).StructProperties()...)
		}
	}

	return
}

func (s SchemaRef) HasAdditionalProperties() bool {
	return openapi.SchemaRef(s).HasAdditionalProperties()
}

func (s SchemaRef) AllSchemas() openapi3.SchemaRefs {
	return openapi.SchemaRef(s).AllSchemas()
}
