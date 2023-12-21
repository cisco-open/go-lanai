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

package aws

import (
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
    "github.com/aws/aws-sdk-go-v2/config"
    "go.uber.org/fx"
)

const FxGroup = `aws`

var Module = &bootstrap.Module{
	Name:       "AWS",
	Precedence: bootstrap.AwsPrecedence,
	Options: []fx.Option{
		fx.Provide(BindAwsProperties, ProvideConfigLoader),
	},
}

func FxCustomizerProvider(constructor interface{}) fx.Annotated {
	return fx.Annotated{
		Group:  FxGroup,
		Target: constructor,
	}
}

type CfgLoaderDI struct {
	fx.In
	Properties  Properties
	Customizers []config.LoadOptionsFunc `group:"aws"`
}

func ProvideConfigLoader(di CfgLoaderDI) ConfigLoader {
	return NewConfigLoader(di.Properties, di.Customizers...)
}
