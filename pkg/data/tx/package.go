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

package tx

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

//var logger = log.New("DB.Tx")

var Module = &bootstrap.Module{
	Name:       "DB Tx",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(provideGormTxManager),
		fx.Invoke(setGlobalTxManager),
	},
}

const (
	FxTransactionExecuterOption = "TransactionExecuterOption"
)

type txDI struct {
	fx.In
	UnnamedTx TxManager                   `optional:"true"`
	DB        *gorm.DB                    `optional:"true"`
	Executer  TransactionExecuter         `optional:"true"`
	Options   []TransactionExecuterOption `group:"TransactionExecuterOption"`
}

type txManagerOut struct {
	fx.Out
	Tx     TxManager `name:"tx/TxManager"`
	GormTx GormTxManager
}

func provideGormTxManager(di txDI) txManagerOut {
	// due to limitation of uber/fx, we cannot override provider, which is not good for testing & mocking
	// the workaround is we always use Named Provider as default,
	// then bail the initialization if an Unnamed one is present
	if di.UnnamedTx != nil {
		if override, ok := di.UnnamedTx.(GormTxManager); ok {
			return txManagerOut{Tx: override, GormTx: override}
		} else {
			// we should avoid this path
			return txManagerOut{Tx: di.UnnamedTx, GormTx: gormTxManagerAdapter{TxManager: di.UnnamedTx}}
		}
	}

	if di.DB == nil {
		panic("default GormTxManager requires a *gorm.DB")
	}

	if di.Executer == nil {
		di.Executer = NewDefaultExecuter(di.Options...)
	}
	m := newGormTxManager(di.DB, di.Executer)
	return txManagerOut{
		Tx:     m,
		GormTx: m,
	}
}
