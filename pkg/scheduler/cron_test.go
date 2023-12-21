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

package scheduler

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

func TestCronTask(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestCronWithDaw(), "TestCronWithDaw"),
		test.GomegaSubTest(SubTestCronWithoutDaw(), "TestCronWithoutDaw"),
		test.GomegaSubTest(SubTestCronWithInvalidExpr(), "TestCronWithInvalidExpr"),
	)
}

/************************
	Sub Tests
 ************************/

func SubTestCronWithDaw() test.GomegaSubTestFunc {
	return func(ctx context.Context, _ *testing.T, g *gomega.WithT) {
		// prepare task
		taskDur := 5 * TestTimeUnit
		tf, execCh := TimingNotifyingTask(taskDur, nil)
		defer close(execCh)

		// run task and verify
		canceller, e := Cron("0 0 0 * * 1", tf)
		g.Expect(e).To(Succeed(), "new task shouldn't return error")
		defer canceller.Cancel()

		// verify options
		t := canceller.(*task)
		g.Expect(t.option.mode).To(BeEquivalentTo(ModeDynamic), "mode should be correct")
		g.Expect(t.option.nextFunc).To(Not(BeNil()), "nextFunc shouldn't be nil")

		// verify next func
		now, _ := time.Parse(time.RFC3339, "2021-10-11T15:04:05Z")
		expected, _ := time.Parse(time.RFC3339, "2021-10-18T00:00:00Z")
		next := t.option.nextFunc(now)
		g.Expect(next).To(gomega.BeTemporally("~", expected, time.Second),
			"next func should return date of next Sunday")
	}
}

func SubTestCronWithoutDaw() test.GomegaSubTestFunc {
	return func(ctx context.Context, _ *testing.T, g *gomega.WithT) {
		// prepare task
		taskDur := 5 * TestTimeUnit
		tf, execCh := TimingNotifyingTask(taskDur, nil)
		defer close(execCh)

		// run task and verify
		canceller, e := Cron("0 0 0 1 *", tf)
		g.Expect(e).To(Succeed(), "new task shouldn't return error")
		defer canceller.Cancel()

		// verify options
		t := canceller.(*task)
		g.Expect(t.option.mode).To(BeEquivalentTo(ModeDynamic), "mode should be correct")
		g.Expect(t.option.nextFunc).To(Not(BeNil()), "nextFunc shouldn't be nil")

		// verify next func
		now, _ := time.Parse(time.RFC3339, "2021-10-11T15:04:05Z")
		expected, _ := time.Parse(time.RFC3339, "2021-11-01T00:00:00Z")
		next := t.option.nextFunc(now)
		g.Expect(next).To(gomega.BeTemporally("~", expected, time.Second),
			"next func should return first day of next month")
	}
}

func SubTestCronWithInvalidExpr() test.GomegaSubTestFunc {
	return func(ctx context.Context, _ *testing.T, g *gomega.WithT) {
		// prepare task
		taskDur := 5 * TestTimeUnit
		tf, execCh := TimingNotifyingTask(taskDur, nil)
		defer close(execCh)

		// run task and verify
		_, e := Cron("0 0 1 *", tf)
		g.Expect(e).To(Not(Succeed()), "new task should return error")
	}
}