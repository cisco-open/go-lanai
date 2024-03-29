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

package request_cache

import (
    "encoding/gob"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/security"
    "go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "request_cache",
	Precedence: security.MinSecurityPrecedence + 20, //after session
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(Module)

	GobRegister()
}

func GobRegister() {
	gob.Register((*CachedRequest)(nil))
}

type initDI struct {
	fx.In
	SecRegistrar security.Registrar `optional:"true"`
}

func register(di initDI) {
	if di.SecRegistrar != nil {
		configurer := newConfigurer()
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
	}
}
