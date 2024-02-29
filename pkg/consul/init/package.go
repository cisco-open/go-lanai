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

package consul

import (
	"embed"
	consulappconfig "github.com/cisco-open/go-lanai/pkg/consul/appconfig"

	"github.com/cisco-open/go-lanai/pkg/appconfig"
	appconfigInit "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/consul"
	consulhealth "github.com/cisco-open/go-lanai/pkg/consul/health"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

//go:embed defaults-consul.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name:       "consul",
	Precedence: bootstrap.ConsulPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(BindConnectionProperties),
		fx.Provide(ProvideDefaultClient),
	},
	Options: []fx.Option{
		appconfigInit.FxEmbeddedDefaults(defaultConfigFS),
		fx.Invoke(consulhealth.Register),
	},
	Modules: []*bootstrap.Module{
		consulappconfig.Module,
	},
}

func Use() {
	bootstrap.Register(Module)
}

func BindConnectionProperties(bootstrapConfig *appconfig.BootstrapConfig) consul.ConnectionProperties {
	c := consul.ConnectionProperties{}
	if e := bootstrapConfig.Bind(&c, consul.PropertyPrefix); e != nil {
		panic(errors.Wrap(e, "failed to bind consul's ConnectionProperties"))
	}
	return c
}

type clientDI struct {
	fx.In
	Props       consul.ConnectionProperties
	Customizers []consul.Options `group:"consul"`
}

func ProvideDefaultClient(di clientDI) (*consul.Connection, error) {
	opts := append([]consul.Options{consul.WithProperties(di.Props)}, di.Customizers...)
	return consul.New(opts...)
}
