package util

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"strings"
	"text/template"
)

var logger = log.New("Internal")
var FuncMap = template.FuncMap{
	"toTitle":       ToTitle,
	"toLower":       ToLower,
	"concat":        Concat,
	"basePath":      BasePath,
	"hasPrefix":     strings.HasPrefix,
	"replaceDashes": ReplaceDash,
	"args":          args,
	"increment":     increment,
	"log":           templateLog,
	"listContains":  ListContains,
	"derefBoolPtr":  derefBoolPtr,
}
