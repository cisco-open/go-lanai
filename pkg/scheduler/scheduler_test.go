package scheduler

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
var TestHook = &TestTaskHook{}

func TestMain(m *testing.M) {
	// This package's tests are time-sensitive. We first need to know the executing host's base performance
	t := time.Now()
	for i := 0; i < 50000; i++ {
		v := time.Now()
		ch := make(chan time.Time, 1)
		ch <- v
		v = <-ch
	}
	TestTimeUnit = (8 * time.Since(t)).Round(time.Millisecond)
	if TestTimeUnit < 8*time.Millisecond {
		TestTimeUnit = 8 * time.Millisecond
	}
	fmt.Printf("Use base TimeUnit [%v] for testing\n", TestTimeUnit)
	AddDefaultHook(TestHook)
	m.Run()
}

/************************
	Tests
 ************************/

func TestTaskTiming(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*TestTimeUnit)
	defer cancel()
	test.RunTest(ctx, t,
		test.GomegaSubTest(SubTestFixedRateTiming(), "TestFixedRate"),
		test.GomegaSubTest(SubTestFixedDelayTiming(), "TestFixedDelay"),
		test.GomegaSubTest(SubTestRunOnceTiming(), "TestRunOnce"),
		test.GomegaSubTest(SubTestStartNowTiming(), "TestStartNow"),
		test.GomegaSubTest(SubTestFixedRateWithPastStartTiming(), "TestFixedRateWithPastStart"),
		test.GomegaSubTest(SubTestFixedDelayWithPastStartTiming(), "TestFixedDelayWithPastStart"),
		test.GomegaSubTest(SubTestDynamicTiming(), "TestDynamicNext"),
	)
}

func TestTaskCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*TestTimeUnit)
	defer cancel()
	test.RunTest(ctx, t,
		test.GomegaSubTest(SubTestCancelOnError(), "TestCancelOnError"),
		test.GomegaSubTest(SubTestCancelOnPanic(), "TestCancelOnPanic"),
		test.GomegaSubTest(SubTestManualCancel(), "TestManual"),
	)
}

func TestTaskSchedulingError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*TestTimeUnit)
	defer cancel()
	test.RunTest(ctx, t,
		test.GomegaSubTest(SubTestSchedulingErrors(), "TestSchedulingError"),
	)
}

/************************
	Sub Tests
 ************************/

func SubTestFixedRateTiming() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const count = 3
		// prepare task
		after := 25 * TestTimeUnit
		rate := 30 * TestTimeUnit
		taskDur := 35 * TestTimeUnit
		tf, execCh := TimingNotifyingTask(taskDur, nil)
		defer close(execCh)

		// run task and verify
		canceller, e := Repeat(tf, StartAfter(after), AtRate(rate), Name("test-fixed-rate"))
		g.Expect(e).To(Succeed(), "new task shouldn't return error")
		defer canceller.Cancel()

		// wait and verify
		start := canceller.(*task).option.initialTime
		check := func(_ TaskCanceller, i int, triggerTime time.Time) {
			expected := start.Add(time.Duration(i * int(rate)))
			g.Expect(triggerTime).To(gomega.BeTemporally("~", expected, TestTimeUnit),
				"task triggered time [i=%d] should be correct", i)
		}
		i, e := WaitTask(ctx, canceller, count, execCh, check)
		g.Expect(i).To(Equal(count), "task should be triggered at least %d times", count)
		g.Expect(e).To(BeNil(), "task shouldn't finished with error")
		g.Expect(TestHook.beforeCount).To(BeNumerically(">", 0), "task before hook should be called")
		g.Expect(TestHook.afterCount).To(BeNumerically(">", 0), "task after hook should be called")
	}
}

func SubTestFixedDelayTiming() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const count = 3
		// prepare task
		after := 25 * TestTimeUnit
		delay := 5 * TestTimeUnit
		taskDur := 25 * TestTimeUnit

		tf, execCh := TimingNotifyingTask(taskDur, nil)
		defer close(execCh)

		// run task and verify
		canceller, e := Repeat(tf, StartAfter(after), WithDelay(delay), Name("test-fixed-delay"))
		g.Expect(e).To(Succeed(), "new task shouldn't return error")
		defer canceller.Cancel()

		// wait and verify
		expected := canceller.(*task).option.initialTime
		check := func(_ TaskCanceller, i int, triggerTime time.Time) {
			g.Expect(triggerTime).To(gomega.BeTemporally("~", expected, 2*TestTimeUnit),
				"task triggered time [i=%d] should be correct", i)
			expected = expected.Add(taskDur).Add(delay)
		}
		i, e := WaitTask(ctx, canceller, count, execCh, check)
		g.Expect(i).To(Equal(count), "task should be triggered at least %d times", count)
		g.Expect(e).To(BeNil(), "task shouldn't finished with error")
		g.Expect(TestHook.beforeCount).To(BeNumerically(">", 0), "task before hook should be called")
		g.Expect(TestHook.afterCount).To(BeNumerically(">", 0), "task after hook should be called")
	}
}

func SubTestRunOnceTiming() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// prepare task
		after := 25 * TestTimeUnit
		taskDur := 5 * TestTimeUnit
		tf, execCh := TimingNotifyingTask(taskDur, nil)
		defer close(execCh)

		// run task and verify
		canceller, e := RunOnce(tf, StartAfter(after), Name("test-run-once"))
		g.Expect(e).To(Succeed(), "new task shouldn't return error")
		defer canceller.Cancel()

		// wait and verify
		start := canceller.(*task).option.initialTime
		check := func(_ TaskCanceller, _ int, triggerTime time.Time) {
			g.Expect(triggerTime).To(gomega.BeTemporally("~", start, TestTimeUnit),
				"task triggered time should be correct")
		}
		i, e := WaitTask(ctx, canceller, 3, execCh, check)
		g.Expect(i).To(Equal(1), "task should be triggered only once")
		g.Expect(e).To(BeNil(), "task shouldn't finished with error")
		g.Expect(TestHook.beforeCount).To(BeNumerically(">", 0), "task before hook should be called")
		g.Expect(TestHook.afterCount).To(BeNumerically(">", 0), "task after hook should be called")
	}
}

func SubTestStartNowTiming() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// prepare task
		taskDur := 5 * TestTimeUnit
		tf, execCh := TimingNotifyingTask(taskDur, nil)
		defer close(execCh)

		// run task and verify
		canceller, e := RunOnce(tf, Name("test-start-now"))
		g.Expect(e).To(Succeed(), "new task shouldn't return error")
		defer canceller.Cancel()

		// wait and verify
		now := time.Now()
		check := func(_ TaskCanceller, _ int, triggerTime time.Time) {
			g.Expect(triggerTime).To(gomega.BeTemporally("~", now, TestTimeUnit),
				"task triggered time should be correct")
		}
		i, e := WaitTask(ctx, canceller, 3, execCh, check)
		g.Expect(i).To(Equal(1), "task should be triggered only once")
		g.Expect(e).To(BeNil(), "task shouldn't finished with error")
		g.Expect(TestHook.beforeCount).To(BeNumerically(">", 0), "task before hook should be called")
		g.Expect(TestHook.afterCount).To(BeNumerically(">", 0), "task after hook should be called")
	}
}

// SubTestFixedRateWithPastStartTiming
// we expect fixed-rate mode respect the past start time and adjust the first execution time accordingly
func SubTestFixedRateWithPastStartTiming() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// prepare task
		const count = 2
		start := time.Now().Add(-time.Second)
		rate := 30 * TestTimeUnit
		taskDur := 35 * TestTimeUnit
		tf, execCh := TimingNotifyingTask(taskDur, nil)
		defer close(execCh)

		// run task and verify
		now := time.Now()
		canceller, e := Repeat(tf, StartAt(start), AtRate(rate), Name("test-fixed-rate"))
		g.Expect(e).To(Succeed(), "new task shouldn't return error")
		defer canceller.Cancel()

		// wait and verify
		for ; !start.After(now); start = start.Add(rate) {
		}
		check := func(_ TaskCanceller, i int, triggerTime time.Time) {
			expected := start.Add(time.Duration(i * int(rate)))
			g.Expect(triggerTime).To(gomega.BeTemporally("~", expected, TestTimeUnit),
				"task triggered time [i=%d] should be correct", i)
		}
		i, e := WaitTask(ctx, canceller, count, execCh, check)
		g.Expect(i).To(Equal(count), "task should be triggered at least %d times", count)
		g.Expect(e).To(BeNil(), "task shouldn't finished with error")
		g.Expect(TestHook.beforeCount).To(BeNumerically(">", 0), "task before hook should be called")
		g.Expect(TestHook.afterCount).To(BeNumerically(">", 0), "task after hook should be called")
	}
}

// SubTestFixedDelayWithPastStartTiming
// we expect fixed-delay mode treat the past start time as start now
func SubTestFixedDelayWithPastStartTiming() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// prepare task
		const count = 2
		start := time.Now().Add(-time.Second)
		rate := 5 * TestTimeUnit
		taskDur := 35 * TestTimeUnit
		tf, execCh := TimingNotifyingTask(taskDur, nil)
		defer close(execCh)

		// run task and verify
		now := time.Now()
		canceller, e := Repeat(tf, StartAt(start), WithDelay(rate), Name("test-fixed-delay"))
		g.Expect(e).To(Succeed(), "new task shouldn't return error")
		defer canceller.Cancel()

		// wait and verify
		check := func(_ TaskCanceller, i int, triggerTime time.Time) {
			if i > 0 {
				return
			}
			g.Expect(triggerTime).To(gomega.BeTemporally("~", now, TestTimeUnit),
				"task triggered time [i=%d] should be correct", i)
		}
		i, e := WaitTask(ctx, canceller, count, execCh, check)
		g.Expect(i).To(Equal(count), "task should be triggered at least %d times", count)
		g.Expect(e).To(BeNil(), "task shouldn't finished with error")
		g.Expect(TestHook.beforeCount).To(BeNumerically(">", 0), "task before hook should be called")
		g.Expect(TestHook.afterCount).To(BeNumerically(">", 0), "task after hook should be called")
	}
}

func SubTestDynamicTiming() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const count = 3

		// prepare task
		timing := []time.Duration{15 * TestTimeUnit, 25 * TestTimeUnit,}
		taskDur := 25 * TestTimeUnit

		tf, execCh := TimingNotifyingTask(taskDur, nil)
		defer close(execCh)

		// run task and verify
		canceller, e := Repeat(tf, dynamicNext(dynamicNextFunc(timing)), Name("test-dynamic"))
		g.Expect(e).To(Succeed(), "new task shouldn't return error")
		defer canceller.Cancel()

		// wait and verify
		expected := time.Now()
		check := func(_ TaskCanceller, i int, triggerTime time.Time) {
			expected = expected.Add(timing[i % len(timing)])
			g.Expect(triggerTime).To(gomega.BeTemporally("~", expected, TestTimeUnit),
				"task triggered time [i=%d] should be correct", i)
		}
		i, e := WaitTask(ctx, canceller, count, execCh, check)
		g.Expect(i).To(Equal(count), "task should be triggered at least %d times", count)
		g.Expect(e).To(BeNil(), "task shouldn't finished with error")
		g.Expect(TestHook.beforeCount).To(BeNumerically(">", 0), "task before hook should be called")
		g.Expect(TestHook.afterCount).To(BeNumerically(">", 0), "task after hook should be called")
	}
}

func SubTestCancelOnError() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const count = 5
		const errAfter = 1

		// prepare task
		after := 25 * TestTimeUnit
		rate := 30 * TestTimeUnit
		taskDur := TestTimeUnit // short execution for task to react errors faster
		tf, execCh := TimingNotifyingTask(taskDur, TaskErrorAfterN(errAfter))
		defer close(execCh)

		// run task and verify
		canceller, e := Repeat(tf, StartAfter(after), AtRate(rate), CancelOnError(), Name("test-cancel-on-error"))
		g.Expect(e).To(Succeed(), "new task shouldn't return error")
		defer canceller.Cancel()

		// wait and verify
		i, e := WaitTask(ctx, canceller, count, execCh, nil)
		g.Expect(i).To(Equal(errAfter+1), "task should be triggered %d times", errAfter+1)
		g.Expect(e).To(Equal(MockedErr), "task should finished with error")
		g.Expect(TestHook.beforeCount).To(BeNumerically(">", 0), "task before hook should be called")
		g.Expect(TestHook.afterCount).To(BeNumerically(">", 0), "task after hook should be called")
	}
}

func SubTestCancelOnPanic() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const count = 5
		const errAfter = 1

		// prepare task
		after := 25 * TestTimeUnit
		rate := 30 * TestTimeUnit
		taskDur := TestTimeUnit // short execution for task to react errors faster
		tf, execCh := TimingNotifyingTask(taskDur, TaskPanicAfterN(errAfter))
		defer close(execCh)

		// run task and verify
		canceller, e := Repeat(tf, StartAfter(after), AtRate(rate), CancelOnError(), Name("test-cancel-on-panic"))
		g.Expect(e).To(Succeed(), "new task shouldn't return error")
		defer canceller.Cancel()

		// wait and verify
		i, e := WaitTask(ctx, canceller, count, execCh, nil)
		g.Expect(i).To(Equal(errAfter+1), "task should be triggered %d times", errAfter+1)
		g.Expect(e).To(Equal(MockedErr), "task should finished with error")
		g.Expect(TestHook.beforeCount).To(BeNumerically(">", 0), "task before hook should be called")
		g.Expect(TestHook.afterCount).To(BeNumerically(">", 0), "task after hook should be called")
	}
}

func SubTestManualCancel() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const count = 5
		const cancelAfter = 1

		// prepare task
		after := 25 * TestTimeUnit
		rate := 30 * TestTimeUnit
		taskDur := TestTimeUnit // short execution for task to react errors faster
		tf, execCh := TimingNotifyingTask(taskDur, nil)
		defer close(execCh)

		// run task and verify
		canceller, e := Repeat(tf, StartAfter(after), AtRate(rate), Name("test-cancel-after-N"))
		g.Expect(e).To(Succeed(), "new task shouldn't return error")
		defer canceller.Cancel()

		// wait, cancel and verify
		check := func(canceller TaskCanceller, i int, _ time.Time) {
			if i >= cancelAfter {
				canceller.Cancel()
			}
		}
		i, e := WaitTask(ctx, canceller, count, execCh, check)
		g.Expect(i).To(Equal(cancelAfter+1), "task should be triggered %d times", cancelAfter+1)
		g.Expect(e).To(Equal(context.Canceled), "task should finished with error")
		g.Expect(TestHook.beforeCount).To(BeNumerically(">", 0), "task before hook should be called")
		g.Expect(TestHook.afterCount).To(BeNumerically(">", 0), "task after hook should be called")
	}
}

func SubTestSchedulingErrors() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// nil task func
		rate := 30 * TestTimeUnit
		_, e = Repeat(nil, AtRate(rate))
		g.Expect(e).To(Not(Succeed()), "Repeat with nil func should fail")

		// repeat without interval
		tf, execCh := TimingNotifyingTask(time.Millisecond, nil)
		defer close(execCh)
		_, e = Repeat(tf)
		g.Expect(e).To(Not(Succeed()), "Repeat without interval should fail")

		// repeat with negative interval
		_, e = Repeat(tf, AtRate(-time.Second))
		g.Expect(e).To(Not(Succeed()), "Repeat with negative interval should fail")

		// StartAfter with negative duration
		_, e = RunOnce(tf, StartAfter(-time.Second))
		g.Expect(e).To(Not(Succeed()), "StartAfter with negative value should fail")
	}
}

/************************
	Helpers
 ************************/

// WaitTask wait given task to be triggered in "count" times or cancelled
func WaitTask(ctx context.Context, canceller TaskCanceller, count int, execCh <-chan time.Time, checkFn TimingCheckFunc) (int, error) {
	for i := 0; i < count; {
		select {
		case triggerTime := <-execCh:
			fmt.Printf("Triggered at %v\n", triggerTime)
			if checkFn != nil {
				checkFn(canceller, i, triggerTime)
			}
			i++
		case e := <-canceller.Cancelled():
			return i, e
		case <-ctx.Done():
			return i, context.DeadlineExceeded
		}
	}
	return count, nil
}

/************************
	Tasks
 ************************/

type TimingCheckFunc func(canceller TaskCanceller, i int, triggerTime time.Time)

type TaskErrorFunc func(i int) error

func TimingNotifyingTask(taskLength time.Duration, errorFn TaskErrorFunc) (TaskFunc, chan time.Time) {
	var i int
	ch := make(chan time.Time, 10)
	return func(ctx context.Context) (err error) {
		ch <- time.Now()
		if errorFn != nil {
			err = errorFn(i)
		}
		i++

		select {
		case <-time.After(taskLength):
		case <-ctx.Done():
		}
		return
	}, ch
}

func dynamicNextFunc(delays []time.Duration) nextFunc {
	var i int
	return func(t time.Time) (next time.Time) {
		d := delays[i % len(delays)]
		next = t.Add(d)
		i++
		return
	}
}

var MockedErr = fmt.Errorf("oops")

func TaskErrorAfterN(n int) TaskErrorFunc {
	return func(i int) error {
		if i >= n {
			return MockedErr
		}
		return nil
	}
}

func TaskPanicAfterN(n int) TaskErrorFunc {
	return func(i int) error {
		if i >= n {
			panic(MockedErr)
		}
		return nil
	}
}

type TestTaskHook struct {
	beforeCount int
	afterCount  int
}

func (h *TestTaskHook) BeforeTrigger(ctx context.Context, _ string) context.Context {
	h.beforeCount++
	return ctx
}

func (h *TestTaskHook) AfterTrigger(_ context.Context, _ string, _ error) {
	h.afterCount++
	return
}
