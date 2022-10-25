package internal

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"text/template"
)

var (
	predefinedRegexes = map[string]string{
		"uuid":      "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$",
		"date":      "^\\\\d{4}-\\\\d{2}-\\\\d{2}$",
		"date-time": "^\\\\d{4}-\\\\d{2}-\\\\d{2}T\\\\d{2}:\\\\d{2}:\\\\d{2}(?:\\\\.\\\\d+)?(?:Z|[\\\\+-]\\\\d{2}:\\\\d{2})?$",
	}

	validatedRegexes = make(map[string]string)

	PackageFuncMap = template.FuncMap{
		"isString":      isString,
		"registerRegex": registerRegex,
		"getRegexName":  getRegexName,
	}
)

func isString(value openapi3.Schema) bool {
	return value.Type == "string"
}
func registerRegex(value openapi3.Schema) string {
	// figure out regex
	name := value.Format
	regexName := predefinedRegexes[name]
	if regexName == "" {
		name = fmt.Sprintf("format%d", len(validatedRegexes))
		regexName = value.Pattern
		if regexName == "" {
			regexName = value.Format
		}
	}
	if validatedRegexes[regexName] != "" {
		return ""
	} else {
		validatedRegexes[regexName] = name
	}

	return regexName
}
func getRegexName(regex string) string {
	return validatedRegexes[regex]
}
