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
	"github.com/cisco-open/go-lanai/pkg/data/cockroach"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

/*************************
	Tests
 *************************/

type gormDI struct {
	fx.In
	GormDB *gorm.DB
}

func TestGormWithDBPlayback(t *testing.T) {
	di := gormDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithDBPlayback("testdb"),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestGormDialetorValidation(&di, &cockroach.GormDialector{}), "GormDialetorValidation"),
	)
}

func TestNoopGorm(t *testing.T) {
	di := gormDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithNoopMocks(),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestGormDialetorValidation(&di, noopGormDialector{}), "GormDialetorValidation"),
		test.GomegaSubTest(SubTestGormDryRun(&di), "TestGormDryRun"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestGormDialetorValidation(di *gormDI, expected interface{}) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.GormDB).To(Not(BeNil()), "*gorm.DB should not be nil")
		g.Expect(di.GormDB.Dialector).To(Not(BeNil()), "Dialector should not be nil")
		g.Expect(di.GormDB.Dialector).To(BeAssignableToTypeOf(expected), "Dialector should be expected type")
	}
}

func SubTestGormDryRun(di *gormDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		rs = di.GormDB.Create(&Model{Value: "doesn't matter"})
		g.Expect(rs.Error).To(Succeed(), "create should succeed")

		var models []*Model
		rs = di.GormDB.Find(&models)
		g.Expect(rs.Error).To(Succeed(), "find should succeed")
	}
}

type Model struct {
	ID    string `gorm:"primaryKey;"`
	Value string
}
