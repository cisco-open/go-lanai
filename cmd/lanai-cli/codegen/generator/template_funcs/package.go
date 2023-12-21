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

package template_funcs

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/go"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/util"
	"github.com/Masterminds/sprig/v3"
	"text/template"
)

var (
	TemplateFuncMaps []template.FuncMap
)

func init() {
	TemplateFuncMaps = []template.FuncMap{
		sprig.GenericFuncMap(),
		util.FuncMap,
		_go.FuncMap,
		openapi.FuncMap,
		lanai.FuncMap,
	}
}

// Load will reset any global registries used internally
func Load() {
	lanai.Load()
	_go.Load()
}

func AddPredefinedRegexes(initialRegexes map[string]string) {
	lanai.AddPredefinedRegexes(initialRegexes)
}
