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

package info

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/actuator"
	"github.com/cisco-open/go-lanai/pkg/appconfig"
)

const (
	ID                   = "info"
	EnableByDefault      = true
	infoPropertiesPrefix = "info"
)

type Input struct {
	Name string `uri:"name"`
}

type Info map[string]interface{}

// InfoEndpoint implements actuator.Endpoint, actuator.WebEndpoint
//goland:noinspection GoNameStartsWithPackageName
type InfoEndpoint struct {
	actuator.WebEndpointBase
	appConfig appconfig.ConfigAccessor
}

func newEndpoint(di regDI) *InfoEndpoint {
	ep := InfoEndpoint{
		appConfig: di.AppContext.Config().(appconfig.ConfigAccessor),
	}
	ep.WebEndpointBase = actuator.MakeWebEndpointBase(func(opt *actuator.EndpointOption) {
		opt.Id = ID
		opt.Ops = []actuator.Operation{
			actuator.NewReadOperation(ep.Read),
		}
		opt.Properties = &di.MgtProperties.Endpoints
		opt.EnabledByDefault = EnableByDefault
	})
	return &ep
}

// Read never returns error
func (ep *InfoEndpoint) Read(ctx context.Context, input *Input) (interface{}, error) {
	info := Info{}
	if e := ep.appConfig.Bind(&info, infoPropertiesPrefix); e != nil {
		return Info{}, e
	}

	buildInfo := map[string]interface{}{}
	if e := ep.appConfig.Bind(&buildInfo, appconfig.PropertyKeyBuildInfo); e == nil {
		info["build-info"] = buildInfo
	}

	logger.WithContext(ctx).Debugf("info %v", info)

	if input.Name == "" {
		return info, nil
	}

	if v, ok := info[input.Name]; ok {
		return v, nil
	}
	return Info{}, nil
}


