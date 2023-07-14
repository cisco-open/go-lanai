package util

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"path"
	"strings"
)

func ToTitle(val string) string {
	return cases.Title(language.AmericanEnglish, cases.NoLower).String(val)
}

func Concat(values ...string) string {
	return strings.Join(values, "")
}

func ToLower(val string) string {
	return strings.ToLower(val)
}

func BasePath(val string) string {
	if val == "" {
		return val
	}
	return path.Base(val)
}

func ReplaceDash(val string) string {
	return strings.ReplaceAll(val, "-", "_")
}
