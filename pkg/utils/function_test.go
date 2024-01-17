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

package utils

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"errors"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

func TestRecoverableFunc(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestWithErrorReturn(), "TestWithErrorReturn"),
		test.GomegaSubTest(SubTestWithoutErrorReturn(), "TestWithoutErrorReturn"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestWithErrorReturn() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const errMsg = `oops`
		var e error
		e = RecoverableFunc(func() error {
			return errors.New(errMsg)
		})()
		g.Expect(e).To(HaveOccurred(), "function should fail when original function fails")
		g.Expect(e.Error()).To(Equal(errMsg), "error message should be correct")

		e = RecoverableFunc(func() error {
			panic(errMsg)
		})()
		g.Expect(e).To(HaveOccurred(), "function should fail when original function panic")
		g.Expect(e.Error()).To(Equal(errMsg), "error message should be correct")

		e = RecoverableFunc(func() error {
			panic(errors.New(errMsg))
		})()
		g.Expect(e).To(HaveOccurred(), "function should fail when original function panic")
		g.Expect(e.Error()).To(Equal(errMsg), "error message should be correct")

		e = RecoverableFunc(func() error {
			return nil
		})()
		g.Expect(e).To(Succeed(), "function should not fail when original function succeeded")
	}
}

func SubTestWithoutErrorReturn() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const errMsg = `oops`
		var e error
		e = RecoverableFunc(func() {
			panic(errMsg)
		})()
		g.Expect(e).To(HaveOccurred(), "function should fail when original function panic")
		g.Expect(e.Error()).To(Equal(errMsg), "error message should be correct")

		e = RecoverableFunc(func() {
			panic(errors.New(errMsg))
		})()
		g.Expect(e).To(HaveOccurred(), "function should fail when original function panic")
		g.Expect(e.Error()).To(Equal(errMsg), "error message should be correct")

		e = RecoverableFunc(func() {
			return
		})()
		g.Expect(e).To(Succeed(), "function should not fail when original function succeeded")
	}
}