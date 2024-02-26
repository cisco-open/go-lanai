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
    "fmt"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs"
    "github.com/cisco-open/go-lanai/pkg/utils/order"
    "github.com/getkin/kin-openapi/openapi3"
)

/**********************
   Data
**********************/

const (
	KDataOpenAPI = "OpenAPIData"
)

/**********************
   Group
**********************/

const (
	gOrderApiCommon = GroupOrderAPI + iota
	gOrderApiStruct
	gOrderApiController
)

type APIGroup struct {
	Option
}

func (g APIGroup) Order() int {
	return GroupOrderAPI
}

func (g APIGroup) Name() string {
	return "API"
}

func (g APIGroup) CustomizeTemplate() (TemplateOptions, error) {
	return func(opt *TemplateOption) {
		// Note: API related functions are already added by default templates, we only need load it with configuration
		template_funcs.AddPredefinedRegexes(g.Components.Contract.Naming.RegExps)
	}, nil
}

func (g APIGroup) CustomizeData(data GenerationData) error {
	openAPIData, e := openapi3.NewLoader().LoadFromFile(g.Components.Contract.Path)
	if e != nil {
		return fmt.Errorf("error parsing OpenAPI file: %v", e)
	}
	data[KDataOpenAPI] = openAPIData

	pInit := data.ProjectMetadata()
	web := ResolveEnabledLanaiModules(LanaiWeb, LanaiActuator, LanaiSwagger)
	pInit.EnabledModules.Add(web.Values()...)
	return nil
}

func (g APIGroup) Generators(opts ...GeneratorOptions) ([]Generator, error) {
	gOpt := GeneratorOption{}
	for _, fn := range opts {
		fn(&gOpt)
	}

	gens := []Generator{
		newDirectoryGenerator(gOpt, func(opt *DirOption) {
			opt.Matcher = isDir().And(matchPatterns("pkg/api/**", "pkg/controllers/**"))
		}),
		newFileGenerator(gOpt, func(opt *FileOption) {
			opt.Order = gOrderApiCommon
			opt.Prefix = "api-common"
		}),
		newApiGenerator(gOpt, func(opt *ApiOption) {
			opt.Order = gOrderApiController
			opt.Prefix = apiDefaultPrefix
		}),
		newApiGenerator(gOpt, func(opt *ApiOption) {
			opt.Order = gOrderApiStruct
			opt.Prefix = apiStructPrefix
		}),
		newApiVersionGenerator(gOpt),
	}
	order.SortStable(gens, order.UnorderedMiddleCompare)
	return gens, nil
}
