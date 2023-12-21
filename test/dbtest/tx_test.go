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

package dbtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/repo"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

/*************************
	Tests
 *************************/

type txDI struct {
	fx.In
	GormAPI repo.GormApi
}

func TestTxManagerWithDBPlayback(t *testing.T) {
	di := txDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithDBPlayback("testdb"),
		apptest.WithModules(repo.Module),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestTxManagerInterface(), "TxManagerInterface"),
		test.GomegaSubTest(SubTestGormTxManagerInterface(&di), "GormTxManagerInterface"),
		test.GomegaSubTest(SubTestManualTxManagerInterface(), "ManualTxManagerInterface"),
		)
}

func TestNoopTxManager(t *testing.T) {
	di := txDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithNoopMocks(),
		apptest.WithModules(repo.Module),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestTxManagerInterface(), "TxManagerInterface"),
		test.GomegaSubTest(SubTestGormTxManagerInterface(&di), "GormTxManagerInterface"),
		test.GomegaSubTest(SubTestManualTxManagerInterface(), "ManualTxManagerInterface"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestTxManagerInterface() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var executed bool
		dummyErr := fmt.Errorf("oops")
		e := tx.Transaction(ctx, func(ctx context.Context) error {
			g.Expect(ctx).To(Not(BeNil()), "ctx in transaction func shouldn't be nil")
			executed = true
			return dummyErr
		})
		g.Expect(e).To(BeIdenticalTo(dummyErr), "tx.Transaction should return operation error")
		g.Expect(executed).To(BeTrue(), "tx.Transaction should executed the function")
	}
}

func SubTestGormTxManagerInterface(di *txDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var executed bool
		dummyErr := fmt.Errorf("oops")
		e := di.GormAPI.Transaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
			g.Expect(ctx).To(Not(BeNil()), "ctx in transaction func shouldn't be nil")
			g.Expect(tx).To(Not(BeNil()), "db in transaction func shouldn't be nil")
			executed = true
			return dummyErr
		})
		g.Expect(e).To(BeIdenticalTo(dummyErr), "GormApi.Transaction should return operation error")
		g.Expect(executed).To(BeTrue(), "GormApi.Transaction should executed the function")
	}
}

func SubTestManualTxManagerInterface() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		txCtx, e := tx.Begin(ctx)
		g.Expect(e).To(Succeed(), "tx.Begin should return no error")
		g.Expect(txCtx).To(Not(BeNil()), "transaction ctx shouldn't be nil")

		cmtCtx, e := tx.Commit(txCtx)
		g.Expect(e).To(Succeed(), "tx.Commit should return no error")
		g.Expect(cmtCtx).To(BeIdenticalTo(ctx), "ctx after commit should be same as before begin")

		txCtx, _ = tx.Begin(ctx)
		rbCtx, e := tx.Rollback(txCtx)
		g.Expect(e).To(Succeed(), "tx.Rollback should return no error")
		g.Expect(rbCtx).To(BeIdenticalTo(ctx), "ctx after rollback should be same as before begin")

		txCtx, _ = tx.Begin(ctx)
		spCtx, e := tx.SavePoint(txCtx, "my-savepoint")
		g.Expect(e).To(Succeed(), "tx.SavePoint should return no error")
		g.Expect(spCtx).To(Not(BeNil()), "ctx after rollback should be same as before begin")

		rbCtx, e = tx.RollbackTo(spCtx, "my-savepoint")
		g.Expect(e).To(Succeed(), "tx.RollbackTo should return no error")
		g.Expect(rbCtx).To(BeIdenticalTo(spCtx), "ctx after RollbackTo should be same as before SavePoint")
	}
}
