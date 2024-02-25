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
	"github.com/cisco-open/go-lanai/pkg/utils/order"
)

/**********************
   Data
**********************/

const (
	KDataSecurity = "Security"
)

/**********************
   Group
**********************/

type SecurityGroup struct {
	Option
}

func (g SecurityGroup) Order() int {
	return GroupOrderSecurity
}

func (g SecurityGroup) Name() string {
	return "Security"
}

func (g SecurityGroup) CustomizeTemplate() (TemplateOptions, error) {
	return nil, nil
}

func (g SecurityGroup) CustomizeData(data GenerationData) error {
	if !g.isApplicable() {
		return nil
	}
	data[KDataSecurity] = g.Components.Security
	modules := make([]*LanaiModule, 0, 4)
	switch g.Components.Security.Authentication.Method {
	case AuthOAuth2:
		modules = append(modules, LanaiSecurity, LanaiResServer)
	}
	switch g.Components.Security.Access.Preset {
	case AccessPresetOPA:
		modules = append(modules, LanaiOPA)
	}
	sec := ResolveEnabledLanaiModules(modules...)
	data.ProjectMetadata().EnabledModules.Add(sec.Values()...)
	return nil
}

func (g SecurityGroup) Generators(opts ...GeneratorOptions) ([]Generator, error) {
	if !g.isApplicable() {
		return []Generator{}, nil
	}
	gOpt := GeneratorOption{}
	for _, fn := range opts {
		fn(&gOpt)
	}

	// Note: for backward compatibility, Default RegenMode is set to ignore
	gens := []Generator{
		newFileGenerator(gOpt, func(opt *FileOption) {
			opt.DefaultRegenMode = RegenModeIgnore
			opt.Prefix = "security"
		}),
		newDirectoryGenerator(gOpt, func(opt *DirOption) {
			opt.Matcher = isDir().And(matchPatterns("configs/**", "pkg/init/**"))
		}),
	}
	order.SortStable(gens, order.UnorderedMiddleCompare)
	return gens, nil
}

func (g SecurityGroup) isApplicable() bool {
	return len(g.Components.Security.Authentication.Method) != 0 && g.Components.Security.Authentication.Method != AuthNone
}
