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