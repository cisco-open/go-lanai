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
