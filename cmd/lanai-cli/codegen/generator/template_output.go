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

package generator

import (
    "bytes"
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
    "path/filepath"
    "regexp"
    "strings"
    "text/template"
)

/*****************************
	Common Output Resolver
 *****************************/

const (
	outputRegexWithFilePrefix = `^(?:%s\.)(?P<filename>.+)(?:\.tmpl)`
)

// regexOutputResolver resolve output descriptor with following rules:
// 1. Apply generation data substitution on template's path
// 2. Use given regular expression resolve output file name.
//    The regular expression must contain a named capturing group "filename". e.g. (?:project\.)(?P<filename>.+)(?:\.tmpl)
// Note: when regex doesn't contains "filename" group, the 2nd step takes no effect
func regexOutputResolver(regex string) TemplateOutputResolver {
    const filenameGroup = `filename`
    compiled := regexp.MustCompile(regex)
    var filenameIdx int
    for i, n := range compiled.SubexpNames() {
        if n == filenameGroup {
            filenameIdx = i
        }
    }
    return TemplateOutputResolverFunc(func(ctx context.Context, tmplDesc TemplateDescriptor, data GenerationData) (TemplateOutputDescriptor, error) {
        resolvedTmplPath, e := resolvePathWithData(tmplDesc.Path, data)
        if e != nil {
            return TemplateOutputDescriptor{}, e
        }
        dir := resolveOutputDir(resolvedTmplPath)

        filename := filepath.Base(resolvedTmplPath)
        matches := compiled.FindStringSubmatch(filename)
        if filenameIdx != 0 && len(matches) > filenameIdx {
            filename = matches[filenameIdx]
        }

        return TemplateOutputDescriptor{
            Path: filepath.Join(dir, filename),
        }, nil
    })
}

// resolveOutputDir will take a template path and return an absolute path of the output dir.
// e.g. pkg/init/project.package.go.tmpl -> /path/to/output/folder/pkg/init
// Note: srcPath should be always relative to template root
func resolveOutputDir(tmplPath string) string {
    return filepath.Join(cmdutils.GlobalArgs.OutputDir, filepath.Dir(tmplPath))
}

var pathVarRegex = regexp.MustCompile(`@([^.].*)@`)
const pathVarReplacement = `@.${1}@`

// resolvePathWithData take an unresolved path and  apply any @...@ with values stored in given data.
// e.g. cmd/@.Project.Name@/project.main.go.tmpl would be resolved to cmd/testservice/project.main.go.tmpl
func resolvePathWithData(unresolved string, data GenerationData) (string, error) {
    toResolve := pathVarRegex.ReplaceAllString(unresolved, pathVarReplacement)
    toResolve = strings.ReplaceAll(toResolve, `@.@`, `.`)
    pathTmpl, e := template.New("filename").Delims("@", "@").Parse(toResolve)
    if e != nil {
        return "", fmt.Errorf("cannot resolve path: [%s]: %v", unresolved, e)
    }
    var buf bytes.Buffer
    if e := pathTmpl.Execute(&buf, data); e != nil {
        return "", fmt.Errorf("cannot resolve path: [%s]: %v", unresolved, e)
    }
    return buf.String(), nil
}
