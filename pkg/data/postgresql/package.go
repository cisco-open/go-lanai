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

package postgresql

import "go.uber.org/fx"

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/data"
	"github.com/cisco-open/go-lanai/pkg/log"
)

var logger = log.New("postgresql")

var Module = &bootstrap.Module{
	Name:       "postgres-compatible",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(NewGormDialetor,
			pqErrorTranslatorProvider(),
			newAnnotatedGormDbCreator(),
		),
	},
}

func Use() {
	bootstrap.Register(Module)
}

/**************************
	Provider
***************************/

func pqErrorTranslatorProvider() fx.Annotated {
	return fx.Annotated{
		Group:  data.GormConfigurerGroup,
		Target: NewPqErrorTranslator,
	}
}
