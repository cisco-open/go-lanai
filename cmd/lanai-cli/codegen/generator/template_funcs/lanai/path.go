// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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
