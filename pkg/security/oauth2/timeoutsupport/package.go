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

package timeoutsupport

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"go.uber.org/fx"
)

//var logger = log.New("SEC.Timeout")

var Module = &bootstrap.Module{
	Name: "timeout",
	Precedence: security.MinSecurityPrecedence + 10, //same as session. since this package doesn't invoke anything, the precedence has no real effect
	Options: []fx.Option{
		fx.Provide(security.BindTimeoutSupportProperties),
		fx.Provide(provideTimeoutSupport),
	},
}

type timeoutDI struct {

}

func provideTimeoutSupport(ctx *bootstrap.ApplicationContext, cf redis.ClientFactory, prop security.TimeoutSupportProperties) oauth2.TimeoutApplier {
	client, err := cf.New(ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = prop.DbIndex
	})

	if err != nil {
		panic(err)
	}

	support := NewRedisTimeoutApplier(client)
	return support
}

func Use() {
	bootstrap.Register(Module)
}