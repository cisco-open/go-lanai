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

package errorhandling

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/security"
	"go.uber.org/fx"
)

var logger = log.New("SEC.Err")

//goland:noinspection GoNameStartsWithPackageName
var Module = &bootstrap.Module{
	Name: "error handling",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(Module)
}

type initDI struct {
	fx.In
	SecRegistrar security.Registrar `optional:"true"`
}

func register(di initDI) {
	if di.SecRegistrar != nil {
		configurer := newErrorHandlingConfigurer()
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
	}
}
