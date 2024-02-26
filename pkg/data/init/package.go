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

package data

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/data"
	"github.com/cisco-open/go-lanai/pkg/data/repo"
	"github.com/cisco-open/go-lanai/pkg/data/tx"
	"github.com/cisco-open/go-lanai/pkg/data/types/pqcrypt"
	"github.com/cisco-open/go-lanai/pkg/web"
	"go.uber.org/fx"
	"reflect"
)

//var logger = log.New("Data")

var Module = &bootstrap.Module{
	Name:       "DB",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(
			data.NewGorm,
			data.ErrorHandlingGormConfigurer(),
			gormErrTranslatorProvider(),
			transactionMaxRetry(),
		),
		web.FxErrorTranslatorProviders(
			webErrTranslatorProvider(data.NewWebDataErrorTranslator),
		),
	},
}

func Use() {
	bootstrap.Register(Module)
	bootstrap.Register(data.Module)
	bootstrap.Register(tx.Module)
	bootstrap.Register(repo.Module)
	bootstrap.Register(pqcrypt.Module)
}

/**************************
	Provider
***************************/

func webErrTranslatorProvider(provider interface{}) func() web.ErrorTranslator {
	return func() web.ErrorTranslator {
		fnv := reflect.ValueOf(provider)
		ret := fnv.Call(nil)
		return ret[0].Interface().(web.ErrorTranslator)
	}
}

func gormErrTranslatorProvider() fx.Annotated {
	return fx.Annotated{
		Group: data.GormConfigurerGroup,
		Target: func() data.ErrorTranslator {
			return data.NewGormErrorTranslator()
		},
	}
}

func transactionMaxRetry() fx.Annotated {
	return fx.Annotated{
		Group: tx.FxTransactionExecuterOption,
		Target: func(properties data.DataProperties) tx.TransactionExecuterOption {
			return tx.MaxRetries(properties.Transaction.MaxRetry, 0)
		},
	}
}

/**************************
	Initialize
***************************/
