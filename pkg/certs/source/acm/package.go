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

package acmcerts

import (
    "encoding/json"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/service/acm"
    awsclient "github.com/cisco-open/go-lanai/pkg/aws"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/certs"
    certsource "github.com/cisco-open/go-lanai/pkg/certs/source"
    "github.com/cisco-open/go-lanai/pkg/log"
    "go.uber.org/fx"
)

var logger = log.New("Certs.ACM")

const (
	sourceType = certs.SourceACM
)

var Module = &bootstrap.Module{
	Name:       "certs-acm",
	Precedence: bootstrap.TlsConfigPrecedence,
	Options: []fx.Option{
		fx.Provide(FxProvider()),
	},
}

func Use() {
	bootstrap.Register(Module)
}

type factoryDI struct {
	fx.In
	AppCtx          *bootstrap.ApplicationContext
	Props           certs.Properties        `optional:"true"`
	AcmClient       *acm.Client             `optional:"true"`
	AwsConfigLoader awsclient.ConfigLoader `optional:"true"`
}

func FxProvider() fx.Annotated {
	return fx.Annotated{
		Group: certs.FxGroup,
		Target: func(di factoryDI) (certs.SourceFactory, error) {
			var client *acm.Client
			switch {
			case di.AcmClient == nil && di.AwsConfigLoader == nil:
				logger.Warnf(`AWS/ACM certificates source is not supported. Tips: Do not forget to initialize ACM client or AWS config loader.`)
				return nil, nil
			case di.AcmClient != nil:
				client = di.AcmClient
			default:
				cfg, e := di.AwsConfigLoader.Load(di.AppCtx)
				if e != nil {
					return nil, fmt.Errorf(`unable to initialize AWS/ACM certificate source: %v`, e)
				}
				client = acm.NewFromConfig(cfg)
			}

			var rawDefaults json.RawMessage
			if di.Props.Sources != nil {
				rawDefaults, _ = di.Props.Sources[sourceType]
			}
			factory, e := certsource.NewFactory[SourceProperties](sourceType, rawDefaults, func(props SourceProperties) certs.Source {
				return NewAcmProvider(di.AppCtx, client, props)
			})
			if e != nil {
				return nil, fmt.Errorf(`unable to register certificate source type [%s]: %v`, sourceType, e)
			}
			return factory, nil
		},
	}
}
