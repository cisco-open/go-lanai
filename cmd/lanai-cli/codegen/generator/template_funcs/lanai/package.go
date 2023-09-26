package lanai

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"text/template"
)

var logger = log.New("Internal")
var FuncMap = template.FuncMap{
	"importsUsedByPath":   ImportsUsedByPath,
	"containsSingularRef": containsSingularRef,
	"isEmpty":             isEmpty,
	"schemaToText":        SchemaToText,
	"propertyToGoType":    PropertyToGoType,
	"shouldHavePointer":   ShouldHavePointer,
	"structTags":          structTags,
	"registerRegex":       registerRegex,
	"regex":               NewRegex,
	"versionList":         versionList,
	"mappingName":         mappingName,
	"mappingPath":         mappingPath,
	"defaultNameFromPath": defaultNameFromPath,
	"components":          NewComponents,
	"requestBody":         NewRequestBody,
	"operation":           NewOperation,
	"property":            NewProperty,
	"schema":              NewSchema,
}

func Load() {
	validatedRegexes = make(map[string]string)
}
func AddPredefinedRegexes(initialRegexes map[string]string) {
	for key, value := range initialRegexes {
		predefinedRegexes[key] = value
	}
}
