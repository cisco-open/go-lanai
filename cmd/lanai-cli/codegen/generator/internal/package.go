package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"text/template"
)

var (
	TemplateFuncMaps []template.FuncMap
	logger           = log.New("Codegen.generator.internal")
)

func init() {
	TemplateFuncMaps = []template.FuncMap{
		regexFuncMap,
		stringsFuncMap,
		structsFuncMap,
		helperFuncMap,
		pathFuncMap,
	}
}

// Load will reset any global registries used internally
func Load() {
	validatedRegexes = make(map[string]string)
	structRegistry = make(map[string]string)
}

func AddPredefinedRegexes(initialRegexes map[string]string) {
	for key, value := range initialRegexes {
		predefinedRegexes[key] = value
	}
}
