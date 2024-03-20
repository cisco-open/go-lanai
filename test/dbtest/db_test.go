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
	"github.com/cisco-open/go-lanai/pkg/data/tx"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"sync"
	"testing"
)

/*************************
	Models
 *************************/

type Client struct {
	ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
	OAuthClientId string    `gorm:"column:oauth_client_id;not null;"`
}

func (Client) TableName() string {
	return "clients"
}

/*************************
	Test
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		EnableDBRecordMode(),
//	)
//}

type testDI struct {
	fx.In
	DB *gorm.DB `optional:"true"`
}

func TestDBPlayback(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithDBPlayback("testdb"),
		apptest.WithDI(di),
		test.SubTestSetup(SetupTestPrepareData(di)),
		test.GomegaSubTest(SubTestExampleSelect(di), "Select"),
		test.GomegaSubTest(SubTestExampleTxSave(di), "TransactionalSave"),
	)
}

/*************************
	Sub Tests
 *************************/

func SetupTestPrepareData(di *testDI) test.SetupFunc {
	var once sync.Once
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		var err error
		once.Do(func() {
			err = di.DB.WithContext(ctx).Exec(`DROP TABLE IF EXISTS clients`).Error
			if err != nil {
				return
			}
			err = di.DB.WithContext(ctx).AutoMigrate(&Client{})
			if err != nil {
				return
			}
			err = di.DB.WithContext(ctx).Create([]*Client{
				{ID: uuid.MustParse("015831a9-978f-437f-b89c-ee4ad960dcdb"), OAuthClientId: "test-client"},
			}).Error
		})
		return ctx, err
	}
}

func SubTestExampleSelect(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.DB).To(Not(BeNil()), "injected gorm.DB should not be nil")

		// select one
		v := Client{}
		r := di.DB.WithContext(ctx).Model(&Client{}).First(&v)
		g.Expect(r.Error).To(Succeed(), "recorded SQL shouldn't introduce error")
		g.Expect(v.ID).To(Not(Equal(uuid.UUID{})), "model should be loaded by First()")

		// select all
		s := make([]*Client, 0)
		r = di.DB.WithContext(ctx).Model(&Client{}).Find(&s)
		g.Expect(r.Error).To(Succeed(), "recorded SQL shouldn't introduce error")
		g.Expect(s).To(Not(BeEmpty()), "slice should not be empty after Find")
	}
}

func SubTestExampleTxSave(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.DB).To(Not(BeNil()), "injected gorm.DB should not be nil")

		e := tx.Transaction(ctx, func(ctx context.Context) error {
			// select one
			v := Client{}
			r := di.DB.WithContext(ctx).Model(&Client{}).First(&v)
			g.Expect(r.Error).To(Succeed(), "select SQL shouldn't introduce error")
			g.Expect(v.ID).To(Not(Equal(uuid.UUID{})), "model should be loaded by First()")

			// save one
			r = di.DB.WithContext(ctx).Save(&v)
			g.Expect(r.Error).To(Succeed(), "save operation shouldn't introduce error")
			return r.Error
		})
		g.Expect(e).To(Succeed(), "tx.Transaction should return no error")
	}
}
