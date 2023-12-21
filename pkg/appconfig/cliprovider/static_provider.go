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

package cliprovider

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
)

const (
	defaultConfigSearchPath = "configs"
)

type StaticConfigProvider struct {
	appconfig.ProviderMeta
	appName string
}

func NewStaticConfigProvider(order int, execCtx *bootstrap.CliExecContext) *StaticConfigProvider {
	return &StaticConfigProvider{
		ProviderMeta: appconfig.ProviderMeta{
			Precedence: order,
		},
		appName: execCtx.Cmd.Root().Name(),
	}
}

func (p *StaticConfigProvider) Name() string {
	return "default"
}

func (p *StaticConfigProvider) Load(_ context.Context) (err error) {
	defer func(){
		p.Loaded = err == nil
	}()

	settings := map[string]interface{}{}

	// Apply application name, profiles, etc
	settings[appconfig.PropertyKeyApplicationName] = p.appName
	settings[appconfig.PropertyKeyConfigFileSearchPath] = []string{defaultConfigSearchPath}
	settings[appconfig.PropertyKeyBuildInfo] = bootstrap.BuildInfoMap

	// un-flatten
	unFlattened, err := appconfig.UnFlatten(settings)
	if err == nil {
		p.Settings = unFlattened
	}

	return
}


