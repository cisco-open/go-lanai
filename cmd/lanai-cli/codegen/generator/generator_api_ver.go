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
	"context"
	"github.com/getkin/kin-openapi/openapi3"
	"io/fs"
	"sort"
	"text/template"
)

// ApiVersionGenerator generate 1 file per API version, based on OpenAPI specs
type ApiVersionGenerator struct {
	data             map[string]interface{}
	template         *template.Template
	defaultRegenRule RegenMode
	rules            RegenRules
	templateFS       fs.FS
	matcher          TemplateMatcher
	outputResolver   TemplateOutputResolver
}

const versionGeneratorName = "version"

type ApiVerOption struct {
	GeneratorOption
}

func newApiVersionGenerator(gOpt GeneratorOption, opts ...func(option *ApiVerOption)) *ApiVersionGenerator {
	o := &ApiVerOption{
		GeneratorOption: gOpt,
	}
	for _, fn := range opts {
		fn(o)
	}

	return &ApiVersionGenerator{
		data:             o.Data,
		template:         o.Template,
		templateFS:       o.TemplateFS,
		matcher:          isTmplFile().And(matchPatterns("**/version.*.tmpl")),
		outputResolver:   regexOutputResolver(`(?:version\.)(?P<filename>.+)(?:\.tmpl)`),
		defaultRegenRule: o.DefaultRegenMode,
		rules:            o.RegenRules,
	}
}

func (g *ApiVersionGenerator) Generate(ctx context.Context, tmplDesc TemplateDescriptor) error {
	if ok, e := g.matcher.Matches(tmplDesc); e != nil || !ok {
		return e
	}

	// get all versions
	iterateOver := make(map[string][]string)
	for pathName, _ := range g.data[KDataOpenAPI].(*openapi3.T).Paths {
		version := apiVersion(pathName)
		iterateOver[version] = append(iterateOver[version], pathName)
	}

	var toGenerate []GenerationContext
	for version, versionData := range iterateOver {
		data := copyOf(g.data)
		sort.Strings(versionData)
		data["VersionData"] = versionData
		data["Version"] = version

		output, e := g.outputResolver.Resolve(ctx, tmplDesc, data)
		if e != nil {
			return e
		}

		regenRule, err := getApplicableRegenRules(output.Path, g.rules, g.defaultRegenRule)
		if err != nil {
			return err
		}
		toGenerate = append(toGenerate, *NewGenerationContext(
			tmplDesc.Path,
			output.Path,
			regenRule,
			data,
		))
	}

	for _, gc := range toGenerate {
		logger.Debugf("[API] generating %v", gc.filename)
		err := GenerateFileFromTemplate(gc, g.template)
		if err != nil {
			return err
		}
		globalCounter.Record(gc.filename)
	}

	return nil
}
