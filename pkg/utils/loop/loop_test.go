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

package loop

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

var TestTimeUnit time.Duration

func TestMain(m *testing.M) {
	// This package's tests are time-sensitive. We first need to know the executing host's base performance
	t := time.Now()
	for i := 0; i < 10000; i++ {
		v := time.Now()
		ch := make(chan time.Time, 1)
		ch <- v
		v = <-ch
	}
	TestTimeUnit = (5 * time.Since(t)).Round(time.Millisecond)
	if TestTimeUnit < 5*time.Millisecond {
		TestTimeUnit = 5 * time.Millisecond
	}
	fmt.Printf("Use base TimeUnit [%v] for testing\n", TestTimeUnit)
	m.Run()
}

/************************
	Tests
 ************************/

func TestSingleExec(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*TestTimeUnit)
	defer cancel()
	test.RunTest(ctx, t,
		test.GomegaSubTest(SubTestExecOrder(), "TestExecOrder"),
		test.GomegaSubTest(SubTestDo(), "TestDo"),
		test.GomegaSubTest(SubTestDoAndWait(), "TestDoAndWait"),
		test.GomegaSubTest(SubTestLoopCancel(), "TestLoopCancel"),
	)
}

func TestRepeatExec(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*TestTimeUnit)
	defer cancel()
	test.RunTest(ctx, t,
		test.GomegaSubTest(SubTestRepeatWithFixedInterval(), "TestFixedInterval"),
		test.GomegaSubTest(SubTestRepeatWithIntervalFunc(), "TestDynamicInterval"),
		test.GomegaSubTest(SubTestRepeatWithExponentialIntervalOnError(), "TestExponentialIntervalOnError"),
	)
}

/************************
	Sub Tests
 ************************/

// SubTestExecOrder verify that tasks are executed one after another
func SubTestExecOrder() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		l := NewLoop()
		ctx, _ = l.Run(ctx)
		g.Expect(ctx).To(Not(BeNil()), "Run() should return non-nil context")

		// add some tasks
		const count = 3
		dur := 10 * TestTimeUnit
		chs := make([]<-chan error, count)
		var order int
		for i := 0; i < count; i++ {
			var tf TaskFunc
			tf, chs[i] = OrderAwareTaskFunc(&order, i, dur)
			l.Do(tf)
		}

		// wait and verify order
		for i, ch := range chs {
			select {
			case e := <-ch:
				g.Expect(e).To(Succeed(), "task %d was not executed successfully: %v", i, e)
			case <-ctx.Done():
				t.Errorf("tasks are not all executed before test timeout")
			}
		}
	}
}

func SubTestDo() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		l := NewLoop()
		ctx, _ = l.Run(ctx)
		g.Expect(ctx).To(Not(BeNil()), "Run() should return non-nil context")

		// run task
		start := time.Now()
		tf := SimpleTaskFunc("whatever", nil, 20*TestTimeUnit)
		l.Do(tf)

		// verify
		now := time.Now()
		g.Expect(now).To(BeTemporally("~", start, TestTimeUnit), "Do should not wait for long lasting tasks")
	}
}

func SubTestDoAndWait() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		l := NewLoop()
		ctx, _ = l.Run(ctx)
		g.Expect(ctx).To(Not(BeNil()), "Run() should return non-nil context")

		// run task without panic, and verify
		expectedRet := "result"
		expectedErr := fmt.Errorf("oops")
		tf := SimpleTaskFunc(expectedRet, expectedErr, 0)
		ret, e := l.DoAndWait(tf)
		g.Expect(ret).To(Equal(expectedRet), "DoAndWait should return correct result")
		g.Expect(e).To(Equal(expectedErr), "DoAndWait should return correct err")

		// run task with panic, and verify
		tf = PanickingTaskFunc(expectedErr, 0)
		_, e = l.DoAndWait(tf)
		g.Expect(e).To(BeEquivalentTo(expectedErr), "DoAndWait should return correct err")
	}
}

func SubTestLoopCancel() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		l := NewLoop()
		ctx, cancel := l.Run(ctx)
		g.Expect(ctx).To(Not(BeNil()), "Run() should return non-nil context")

		// run task without panic, and verify
		expectedRet := "result"
		tf := SimpleTaskFunc(expectedRet, nil, 20 * TestTimeUnit)
		go func() {
			time.Sleep(2 * TestTimeUnit)
			cancel()
		}()
		ret, e := l.DoAndWait(tf)

		g.Expect(ret).To(BeNil(), "task should be cancelled")
		g.Expect(e).To(Equal(context.Canceled), "task should yield context.Cancelled error")
	}
}

func SubTestRepeatWithFixedInterval() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		l := NewLoop()
		ctx, _ = l.Run(ctx)
		g.Expect(ctx).To(Not(BeNil()), "Run() should return non-nil context")

		// add repeated task
		const count = 3
		interval := 10 * TestTimeUnit
		errFn := func(int) error {return fmt.Errorf("oops")}
		tf, ch := RepeatableTaskFunc("result", errFn, 0)
		l.Repeat(tf, FixedRepeatInterval(interval))

		// verify timing
		expected := time.Now()
		for i := 0; i < count; i++ {
			select {
			case execTime := <-ch:
				g.Expect(execTime).To(BeTemporally("~", expected, TestTimeUnit), "Repeat task %d should be executed at correct time", i)
				expected = expected.Add(interval)
			case <-ctx.Done():
				t.Errorf("tasks are not all executed before test timeout")
			}
		}
	}
}

func SubTestRepeatWithIntervalFunc() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		l := NewLoop()
		ctx, _ = l.Run(ctx)
		g.Expect(ctx).To(Not(BeNil()), "Run() should return non-nil context")

		// add repeated task
		const count = 3
		intervals := []time.Duration{
			8 * TestTimeUnit,
			15 * TestTimeUnit,
		}
		errFn := func(int) error {return nil}
		tf, ch := RepeatableTaskFunc("result", errFn, 0)
		l.Repeat(tf, func(opt *TaskOption) {
			opt.RepeatIntervalFunc = StaticRepeatIntervalFunc(intervals)
		})

		// verify timing
		expected := time.Now()
		for i := 0; i < count; i++ {
			select {
			case execTime := <-ch:
				g.Expect(execTime).To(BeTemporally("~", expected, TestTimeUnit), "Repeat task %d should be executed at correct time", i)
				expected = expected.Add(intervals[i % len(intervals)])
			case <-ctx.Done():
				t.Errorf("tasks are not all executed before test timeout")
			}
		}
	}
}

func SubTestRepeatWithExponentialIntervalOnError() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		l := NewLoop()
		ctx, _ = l.Run(ctx)
		g.Expect(ctx).To(Not(BeNil()), "Run() should return non-nil context")

		// add repeated task
		const count = 3
		initInterval := 8 * TestTimeUnit
		factor := 2.0
		errFn := func(i int) error {
			// error on even number of iteration
			if i % 2 == 0 {
				return fmt.Errorf("oops")
			}
			return nil
		}
		tf, ch := RepeatableTaskFunc("result", errFn, 0)
		l.Repeat(tf, ExponentialRepeatIntervalOnError(initInterval, factor))

		// verify timing
		expected := time.Now()
		for i := 0; i < count; i++ {
			select {
			case execTime := <-ch:
				g.Expect(execTime).To(BeTemporally("~", expected, TestTimeUnit), "Repeat task %d should be executed at correct time", i)
				if i % 2 == 0 {
					expected = expected.Add(time.Duration(factor) * initInterval)
				} else {expected = expected.Add(initInterval)

				}
			case <-ctx.Done():
				t.Errorf("tasks are not all executed before test timeout")
			}
		}
	}
}

/************************
	Helpers
 ************************/

func OrderAwareTaskFunc(order *int, expectedOrder int, dur time.Duration) (TaskFunc, <-chan error) {
	ch := make(chan error, 1)
	return func(ctx context.Context, l *Loop) (ret interface{}, err error) {
		defer func() {
			ch <- err
			close(ch)
		}()
		if expectedOrder != *order {
			return *order, fmt.Errorf("incorrect exec order")
		}
		select {
		case <-time.After(dur):
			*order = *order + 1
		case <-ctx.Done():
		}
		return *order, nil
	}, ch
}

func SimpleTaskFunc(ret interface{}, err error, dur time.Duration) TaskFunc {
	return func(ctx context.Context, l *Loop) (interface{}, error) {
		if dur != 0 {
			select {
			case <-time.After(dur):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return ret, err
	}
}

func PanickingTaskFunc(err error, dur time.Duration) TaskFunc {
	return func(ctx context.Context, l *Loop) (interface{}, error) {
		if dur != 0 {
			select {
			case <-time.After(dur):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		panic(err)
	}
}

func RepeatableTaskFunc(ret interface{}, errFn func(int) error, dur time.Duration) (TaskFunc, <-chan time.Time) {
	var i int
	ch := make(chan time.Time, 1)
	return func(ctx context.Context, l *Loop) (interface{}, error) {
		defer func() {
			i++
		}()
		ch <- time.Now()
		if dur != 0 {
			select {
			case <-time.After(dur):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return ret, errFn(i)
	}, ch
}

func StaticRepeatIntervalFunc(intervals []time.Duration) RepeatIntervalFunc {
	var i int
	return func(_ interface{}, _ error) (ret time.Duration) {
		ret = intervals[i % len(intervals)]
		i++
		return
	}
}
