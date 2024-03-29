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

package apilist

import (
	"github.com/cisco-open/go-lanai/pkg/actuator"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/log"
	"go.uber.org/fx"
	"io/fs"
	"os"
)

var logger = log.New("ACTR.APIList")

var staticFS = []fs.FS{os.DirFS(".")}

var Module = &bootstrap.Module{
	Name:       "actuator-apilist",
	Precedence: actuator.MinActuatorPrecedence,
	Options: []fx.Option{
		fx.Provide(BindProperties),
		fx.Invoke(register),
	},
}

func Register() {
	bootstrap.Register(Module)
}

func StaticFS(fs ...fs.FS) {
	if len(fs) != 0 {
		staticFS = fs
	}
}

type regDI struct {
	fx.In
	Registrar     *actuator.Registrar
	MgtProperties actuator.ManagementProperties
	Properties    Properties
}

func register(di regDI) {
	ep := newEndpoint(di)
	di.Registrar.MustRegister(ep)
}
