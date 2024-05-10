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

package data_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/data"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/dbtest"
	gomegautils "github.com/cisco-open/go-lanai/test/utils/gomega"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

var (
	TestModelID1   = uuid.MustParse("92d22359-6e61-4407-adf1-cee2ae8b8262")
	TestModelID2   = uuid.MustParse("63299139-748d-44bc-bf9a-cdd79389ad68")
	TestModelID3   = uuid.MustParse("3468d223-11ed-4dfc-9e50-fd385dc57099")
	PreparedModels = map[string]uuid.UUID{
		"Model-1": TestModelID1,
	}
)

/*************************
	Test
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

type errTestDI struct {
	fx.In
	dbtest.DI
}

func TestErrorTranslation(t *testing.T) {
	di := &errTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		//apptest.WithTimeout(60*time.Minute),
		apptest.WithDI(di),
		test.SubTestSetup(SetupWithTable(&di.DI)),
		test.SubTestTeardown(TeardownWithTruncateTable(&di.DI)),
		test.GomegaSubTest(SubTestServerSideErrorTranslation(di), "ServerSideErrorTranslation"),
	)
}

/*************************
	Test
 *************************/

func SetupWithTable(di *dbtest.DI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		g := gomega.NewWithT(t)
		r := di.DB.Exec(tableSQL)
		g.Expect(r.Error).To(Succeed(), "create table shouldn't fail")
		r = di.DB.Exec(fmt.Sprintf(`TRUNCATE TABLE "%s" RESTRICT`, TestModel{}.TableName()))
		g.Expect(r.Error).To(Succeed(), "truncate table shouldn't fail")
		for k, v := range PreparedModels {
			m := TestModel{
				ID:        v,
				UniqueKey: k,
				Value:     fmt.Sprintf("Value of %s", k),
			}
			r = di.DB.Create(&m)
			g.Expect(r.Error).To(Succeed(), "create model [%s] shouldn't fail", k)
		}
		return ctx, nil
	}
}

func TeardownWithTruncateTable(di *dbtest.DI) test.TeardownFunc {
	return func(ctx context.Context, t *testing.T) error {
		return nil
	}
}

func SubTestServerSideErrorTranslation(di *errTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// duplicated key
		m := TestModel{
			ID:        TestModelID2,
			UniqueKey: "Model-1", // duplicated key
			Value:     "what ever",
		}
		expected := data.NewDuplicateKeyError("mocked error")
		r := di.DB.Create(&m)
		g.Expect(r.Error).To(HaveOccurred(), "create model [%s] should fail", m.UniqueKey)
		g.Expect(r.Error).To(gomegautils.IsError(expected), "error should be correct type")
		var dataE data.DataError
		ok := errors.As(r.Error, &dataE)
		g.Expect(ok).To(BeTrue(), "error should be data.DataError type")
		g.Expect(dataE.RootCause()).To(BeAssignableToTypeOf(&pq.Error{}), "error should have cause with pq.Error type")
		g.Expect(dataE.Error()).To(ContainSubstring("Model-1"), "error message should contain the duplicated key")
		g.Expect(dataE.Error()).To(ContainSubstring("uk"), "error message should contain the column name")

		// record not found
		// expect the error message to be different
		dest := TestModel{}
		r = di.DB.First(&dest, TestModel{UniqueKey: "Non-Exist-Model-Number"})
		g.Expect(r.Error).To(HaveOccurred())
		g.Expect(r.Error).To(gomegautils.IsError(data.ErrorRecordNotFound), "error should be record not found")
		ok = errors.As(r.Error, &dataE)
		g.Expect(dataE.Error()).To(ContainSubstring("record not found"), "error message should contain record not found")
	}
}

/*************************
	Mocks
 *************************/

type TestModel struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	UniqueKey string    `gorm:"uniqueIndex;column:uk"`
	Value     string
}

func (TestModel) TableName() string {
	return "test_model"
}

const tableSQL = `
CREATE TABLE IF NOT EXISTS public.test_model (
	id UUID NOT NULL DEFAULT gen_random_uuid(),
	"uk" STRING NOT NULL,
	"value" STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY (id ASC),
	UNIQUE INDEX idx_unique_key (uk ASC),
	FAMILY "primary" (id, uk, value)
);`
