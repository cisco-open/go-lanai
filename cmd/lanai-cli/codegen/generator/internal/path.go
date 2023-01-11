package internal

import (
	"github.com/getkin/kin-openapi/openapi3"
	"regexp"
	"sort"
	"text/template"
)

var (
	pathFuncMap = template.FuncMap{
		"versionList": versionList,
	}
)

func versionList(paths openapi3.Paths) []string {
	var result []string
	for p, _ := range paths {
		parts := regexp.MustCompile(".+\\/(v\\d+)\\/(.+)").FindStringSubmatch(p)
		if len(parts) > 2 {
			version := parts[1]
			if !listContains(result, version) {
				result = append(result, version)
			}
		}
	}
	sort.Strings(result)
	return result
}
