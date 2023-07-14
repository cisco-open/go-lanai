package _go

import (
	"text/template"
)

var (
	FuncMap = template.FuncMap{
		"registerStruct": RegisterStruct,
		"structLocation": StructLocation,
		"structRegistry": StructRegistry,
		"NewStruct":      NewStruct,
		"NewMyProperty":  NewProperty,
		"NewImports":     NewImports,
	}
)

func Load() {
	structRegistry = make(map[string]string)
}
