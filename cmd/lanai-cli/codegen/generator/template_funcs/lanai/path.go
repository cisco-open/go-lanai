package lanai

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/util"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"regexp"
	"sort"
	"strings"
)

const (
	PathRegex = ".+(\\/api\\/(v\\d+)\\/(.+))"
)
const (
	FullPath = iota
	PathFromApi
	VersionFromPath
	PathAfterVersion
)

func versionList(paths openapi3.Paths) []string {
	var result []string
	for p, _ := range paths {
		version := pathPart(p, VersionFromPath)
		if !util.ListContains(result, version) {
			result = append(result, version)
		}
	}
	sort.Strings(result)
	return result
}

func mappingName(path, operation string) string {
	result := pathPart(path, PathAfterVersion)
	result = replaceParameterDelimiters(result, "", "")
	result = strings.ReplaceAll(result, "/", "-")

	return strings.ToLower(fmt.Sprintf("%v-%v", result, operation))
}

func mappingPath(path string) (result string) {
	result = pathPart(path, PathFromApi)
	result = replaceParameterDelimiters(result, ":", "")

	return result
}

func defaultNameFromPath(val string) string {
	path := pathPart(val, PathAfterVersion)
	path = replaceParameterDelimiters(path, "/", "")
	pathParts := strings.Split(path, "/")

	// make this camelCase
	for p := range pathParts {
		if p == 0 {
			continue
		}
		pathParts[p] = util.ToTitle(pathParts[p])
	}

	return strings.Join(pathParts, "")
}

func pathPart(path string, pathPart int) (result string) {
	parts := regexp.MustCompile(PathRegex).FindStringSubmatch(path)
	if len(parts) > pathPart {
		result = parts[pathPart]
	}
	return result
}

func replaceParameterDelimiters(path, leftDelim, rightDelim string) (result string) {
	result = strings.ReplaceAll(path, "{", leftDelim)
	result = strings.ReplaceAll(result, "}", rightDelim)
	return result
}
