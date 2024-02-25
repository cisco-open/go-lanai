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
    "github.com/cisco-open/go-lanai/pkg/utils"
    "reflect"
    "strings"
)

/**********************
   Scaffolding Data
**********************/

type LanaiModule struct {
	Name        string
	Alias       string
	InitPackage string
}

func (m LanaiModule) ImportPath(root string) string {
	return fmt.Sprintf(`%s/%s`, root, m.InitPackage)
}

// ImportAlias return alias for importing. Return empty if no alias is needed
func (m LanaiModule) ImportAlias() string {
	if len(m.Alias) != 0 {
		// Alias is specifically set
		return m.Alias
	}
	if strings.HasSuffix(m.InitPackage, m.Name) {
		// package is same name as the name, no alias
		return ""
	}
	return m.Name
}

// Ref returns package reference when using the package in code, always not empty
// The result could be either the name or alias. The behavior should match ImportAlias
func (m LanaiModule) Ref() string {
	if len(m.Alias) != 0 {
		return m.Alias
	}
	return m.Name
}

// Lanai Modules Definitions
var (
	LanaiAppConfig     = &LanaiModule{Name: "appconfig", InitPackage: "appconfig/init"}
	LanaiRedis         = &LanaiModule{Name: "redis", InitPackage: "redis"}
	LanaiSecurity      = &LanaiModule{Name: "security", InitPackage: "security"}
	LanaiResServer     = &LanaiModule{Name: "resserver", InitPackage: "security/config/resserver"}
	LanaiWeb           = &LanaiModule{Name: "web", InitPackage: "web/init"}
	LanaiConsul        = &LanaiModule{Name: "consul", InitPackage: "consul/init"}
	LanaiVault         = &LanaiModule{Name: "vault", InitPackage: "vault/init"}
	LanaiDiscovery     = &LanaiModule{Name: "discovery", InitPackage: "discovery/init"}
	LanaiActuator      = &LanaiModule{Name: "actuator", InitPackage: "actuator/init"}
	LanaiSwagger       = &LanaiModule{Name: "swagger", InitPackage: "swagger"}
	LanaiTracing       = &LanaiModule{Name: "tracing", InitPackage: "tracing/init"}
	LanaiData          = &LanaiModule{Name: "data", InitPackage: "data/init"}
	LanaiCockroach     = &LanaiModule{Name: "cockroach", InitPackage: "cockroach"}
	LanaiKafka         = &LanaiModule{Name: "kafka", InitPackage: "kafka"}
	LanaiHttpClient    = &LanaiModule{Name: "httpclient", InitPackage: "integrate/httpclient"}
	LanaiSecurityScope = &LanaiModule{Name: "scope", InitPackage: "integrate/security/scope"}
	LanaiDSync         = &LanaiModule{Name: "dsync", InitPackage: "dsync"}
	LanaiOPA           = &LanaiModule{Name: "opa", InitPackage: "opa/init", Alias: "opainit"}
	//Lanai   = &LanaiModule{Name: "", InitPackage: "/init"}
)

type LanaiModuleGroup []*LanaiModule

func (g LanaiModuleGroup) FilterByName(filter utils.StringSet) []*LanaiModule {
	ret := make([]*LanaiModule, 0, len(g))
	for _, m := range g {
		if filter.Has(m.Name) {
			ret = append(ret, m)
		}
	}
	return ret
}

type LanaiModules struct {
	Basic       LanaiModuleGroup
	Web         LanaiModuleGroup
	Data        LanaiModuleGroup
	Integration LanaiModuleGroup
	Security    LanaiModuleGroup
	Others      LanaiModuleGroup
}

func (m LanaiModules) Modules() (ret LanaiModuleGroup) {
	ret = make(LanaiModuleGroup, 0, 20)
	rv := reflect.ValueOf(m)
	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i).Interface()
		if modules, ok := f.(LanaiModuleGroup); ok {
			ret = append(ret, modules...)
		}
	}
	return
}

// SupportedLanaiModules is all lanai modules by groups
var SupportedLanaiModules = LanaiModules{
	Basic:       []*LanaiModule{LanaiAppConfig, LanaiConsul, LanaiVault, LanaiRedis, LanaiTracing},
	Web:         []*LanaiModule{LanaiWeb, LanaiActuator, LanaiSwagger},
	Data:        []*LanaiModule{LanaiData, LanaiCockroach},
	Integration: []*LanaiModule{LanaiDiscovery, LanaiHttpClient, LanaiSecurityScope, LanaiKafka},
	Security:    []*LanaiModule{LanaiSecurity, LanaiResServer, LanaiOPA},
	Others:      []*LanaiModule{LanaiDSync},
}

// LanaiModuleDependencies some module requires other modules to be initialized.
// The dependency list doesn't include always-required modules like "bootstrap" and "appconfig"
var LanaiModuleDependencies = map[*LanaiModule][]*LanaiModule{
	LanaiSecurity:      {LanaiRedis},
	LanaiResServer:     {LanaiSecurity},
	LanaiDiscovery:     {LanaiConsul},
	LanaiActuator:      {LanaiWeb},
	LanaiSwagger:       {LanaiWeb},
	LanaiCockroach:     {LanaiData},
	LanaiSecurityScope: {LanaiSecurity},
	LanaiDSync:         {LanaiConsul},
}

// ResolveEnabledLanaiModules resolve all modules required by given modules, based on dependencies
func ResolveEnabledLanaiModules(modules ...*LanaiModule) (ret utils.StringSet) {
	ret = utils.NewStringSet(LanaiAppConfig.Name)
	for _, m := range modules {
		resolveModuleDependencies(m, ret)
	}
	return
}

func resolveModuleDependencies(module *LanaiModule, visited utils.StringSet) []*LanaiModule {
	if visited.Has(module.Name) {
		return []*LanaiModule{}
	}
	visited.Add(module.Name)
	deps, ok := LanaiModuleDependencies[module]
	if !ok || len(deps) == 0 {
		return []*LanaiModule{module}
	}
	ret := make([]*LanaiModule, 1, 20)
	ret[0] = module
	for _, m := range deps {
		ret = append(ret, resolveModuleDependencies(m, visited)...)
	}
	return ret
}
