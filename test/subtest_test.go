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

package test

import (
	"context"
	"github.com/onsi/gomega"
	"testing"
)

func TestWithSubTests(t *testing.T) {
	RunTest(context.Background(), t,
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-1"),
		// will be removed by RemoveSubTests
		GomegaSubTest(SubTestAlwaysFail(), "WillFail"),
		// without specific name
		GomegaSubTest(SubTestAlwaysSucceed()),
		// with name arguments
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest %s", "4"),
		RemoveSubTests("WillFail"),
	)
}

func TestWithMixedResults(t *testing.T) {
	t.Skipf("Skipped because this test is meant to fail")
	RunTest(context.Background(), t,
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-1"),
		GomegaSubTest(SubTestAlwaysFail(), "FailedTest-1"),
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-2"),
		GomegaSubTest(SubTestAlwaysFail(), "FailedTest-2"),
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-3"),
		GomegaSubTest(SubTestAlwaysFail(), "FailedTest-3"),
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-4"),
		GomegaSubTest(SubTestAlwaysFail(), "FailedTest-4"),
	)
}


func SubTestAlwaysSucceed() GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(true).To(gomega.BeTrue())
	}
}

func SubTestAlwaysFail() GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(true).To(gomega.BeFalse())
	}
}

func RemoveSubTests(names...string) Options {
	return func(opt *T) {
		for _, name := range names {
			opt.SubTests.Delete(name)
		}
	}
}