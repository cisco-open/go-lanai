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

const (
	gOrderCleanup = GroupOrderProject + iota
)

/**********************
   Group
**********************/

type ProjectGroup struct {
	Option
}

func (g ProjectGroup) Order() int {
	return GroupOrderProject
}

func (g ProjectGroup) Name() string {
	return "Project"
}

func (g ProjectGroup) CustomizeTemplate() (TemplateOptions, error) {
	return nil, nil
}

func (g ProjectGroup) CustomizeData(data GenerationData) error {
	basic := ResolveEnabledLanaiModules(LanaiAppConfig, LanaiConsul, LanaiVault, LanaiRedis, LanaiTracing, LanaiDiscovery)
	pInit := data.ProjectMetadata()
	pInit.EnabledModules.Add(basic.Values()...)
	return nil
}

func (g ProjectGroup) Generators(opts ...GeneratorOptions) ([]Generator, error) {
	gOpt := GeneratorOption{}
	for _, fn := range opts {
		fn(&gOpt)
	}

	// Note: for backward compatibility, Default RegenMode is set to ignore
	gens := []Generator{
		newFileGenerator(gOpt, func(opt *FileOption) {
			opt.DefaultRegenMode = RegenModeIgnore
			opt.Prefix = "project"
		}),
		newDirectoryGenerator(gOpt, func(opt *DirOption) {
			opt.Matcher = isDir().And(matchPatterns("cmd/**", "configs/**", "pkg/init/**"))
		}),
		newCleanupGenerator(gOpt, func(opt *CleanupOption) {
			opt.Order = gOrderCleanup
		}),
	}
	order.SortStable(gens, order.UnorderedMiddleCompare)
	return gens, nil
}
