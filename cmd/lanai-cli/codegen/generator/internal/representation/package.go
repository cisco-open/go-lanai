package representation

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/getkin/kin-openapi/openapi3"
	"strings"
	"text/template"
)

var (
	logger  = log.New("Codegen.generator.internal.representations")
	FuncMap = template.FuncMap{
		"property":           NewProperty,
		"propertyTypePrefix": PropertyTypePrefix,
		"operation":          NewOperation,
		"schema":             NewSchema,
		"components":         NewComponents,
	}
)

const UUID_IMPORT_PATH = "github.com/google/uuid"

func isUUID(schema *openapi3.SchemaRef) bool {
	if schema.Ref != "" {
		return false
	}
	if schema.Value.Type == "array" {
		return isUUID(schema.Value.Items)
	} else if schema.Value.Type != "string" {
		return false
	}
	return strings.ToLower(schema.Value.Pattern) == "uuid" || strings.ToLower(schema.Value.Format) == "uuid"
}
