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

package profiler

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
)

const (
	RouteGroup      = "debug"
	PathPrefixPProf = "pprof"
)

var Module = &bootstrap.Module{
	Precedence: bootstrap.DebugPrecedence,
	Options: []fx.Option{
		fx.Invoke(initialize),
	},
}

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}

type initDI struct {
	fx.In
	Lifecycle    fx.Lifecycle
	WebRegistrar *web.Registrar `optional:"true"`
}

func initialize(di initDI) {
	if di.WebRegistrar == nil {
		return
	}
	di.WebRegistrar.MustRegister(&PProfController{})
}

