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
    "context"
    "database/sql"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/onsi/gomega"
    "go.uber.org/fx"
    "gorm.io/gorm"
    "testing"
)

type noopTxManager struct {}

func provideNoopTxManager() TxManager {
	return noopTxManager{}
}

func (m noopTxManager) Transaction(_ context.Context, _ TxFunc, _ ...*sql.TxOptions) error {
	return nil
}

func (m noopTxManager) WithDB(_ *gorm.DB) GormTxManager {
	return m
}

/*************************
	Tests
 *************************/

func TestOverridingTxManager(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(Module),
		apptest.WithFxOptions(
			fx.Provide(provideNoopTxManager),
		),
		test.GomegaSubTest(SubTestNoopTxManager(), "VerifyNoopTxManager"),
	)
}

// TODO more tests

/*************************
	Sub Tests
 *************************/

func SubTestNoopTxManager() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(txManager).To(gomega.BeAssignableToTypeOf(noopTxManager{}))
		e := Transaction(context.Background(), func(ctx context.Context) error {
			return nil
		})
		g.Expect(e).To(gomega.Succeed(), "TxManager shouldn't return error")
	}
}