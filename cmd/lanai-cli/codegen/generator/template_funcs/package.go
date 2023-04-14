package template_funcs

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/go"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/util"
	"text/template"
)

var (
	TemplateFuncMaps []template.FuncMap
)

func init() {
	TemplateFuncMaps = []template.FuncMap{
		util.FuncMap,
		_go.FuncMap,
		openapi.FuncMap,
		lanai.FuncMap,
	}
}

// Load will reset any global registries used internally
func Load() {
	lanai.Load()
	_go.Load()
}

func AddPredefinedRegexes(initialRegexes map[string]string) {
	lanai.AddPredefinedRegexes(initialRegexes)
}
