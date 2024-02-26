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

package acm

import (
    "github.com/aws/aws-sdk-go-v2/service/acm"
    awsclient "github.com/cisco-open/go-lanai/pkg/aws"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/log"
    "go.uber.org/fx"
)

var logger = log.New("Aws")

var Module = &bootstrap.Module{
	Name:       "ACM",
	Precedence: bootstrap.AwsPrecedence,
	Options: []fx.Option{
		fx.Provide(NewClientFactory),
		fx.Provide(NewDefaultClient),
		fx.Invoke(RegisterHealth),
	},
}

// Use func, does nothing. Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
	bootstrap.Register(awsclient.Module)
}

func NewDefaultClient(ctx *bootstrap.ApplicationContext, factory ClientFactory) (*acm.Client, error) {
	return factory.New(ctx)
}
