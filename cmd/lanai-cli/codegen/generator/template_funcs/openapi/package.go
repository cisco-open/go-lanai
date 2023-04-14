package openapi

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"text/template"
)

var FuncMap = template.FuncMap{
	"requiredList": requiredList,
}

var logger = log.New("Codegen.generator.internal")
