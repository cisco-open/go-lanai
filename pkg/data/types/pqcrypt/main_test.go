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

package pqcrypt

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Test Cases
 *************************/

// Note: Encrypt and Decrypt are covered in map_test.go

func TestCreateKey(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(Module),
		apptest.WithFxOptions(
			fx.Provide(newMockedEncryptor(true)),
		),
		test.GomegaSubTest(SubTestCreateKey(uuid.New(), true), "CreateKeySuccess"),
		test.GomegaSubTest(SubTestCreateKey(uuid.UUID{}, false), "CreateKeyFail"),
	)
}

func TestNoopCreateKey(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(Module),
		apptest.WithFxOptions(
			fx.Provide(newMockedEncryptor(false)),
		),
		test.GomegaSubTest(SubTestCreateKey(uuid.New(), true), "CreateKeyWithValidKey"),
		test.GomegaSubTest(SubTestCreateKey(uuid.UUID{}, true), "CreateKeyWithInvalidKey"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestCreateKey(uuid uuid.UUID, expectSuccess bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := CreateKeyWithUUID(ctx, uuid)
		if expectSuccess {
			g.Expect(e).To(Succeed(), "CreateKey should success")
		} else {
			g.Expect(e).To(Not(Succeed()), "CreateKey should fail")
		}
	}
}
