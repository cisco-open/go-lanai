package internal

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"path"
	"strings"
	"text/template"
)

var (
	stringsFuncMap = template.FuncMap{
		"toTitle":   toTitle,
		"concat":    concat,
		"basePath":  basePath,
		"toLower":   toLower,
		"hasPrefix": strings.HasPrefix,
	}
)

func toTitle(val string) string {
	return cases.Title(language.AmericanEnglish, cases.NoLower).String(val)
}

func concat(values ...string) string {
	return strings.Join(values, "")
}

func toLower(val string) string {
	return strings.ToLower(val)
}

func basePath(val string) string {
	if val == "" {
		return val
	}
	return path.Base(val)
}
