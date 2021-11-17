package scheduler

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"fmt"
	"time"
)

var logger = log.New("Scheduler")

var defaultTaskHooks []TaskHook

/**************************
	Scheduling Functions
 **************************/

// Repeat schedules a takes that repeat at specified time
func Repeat(taskFunc TaskFunc, opts ...TaskOptions) (TaskCanceller, error) {
	opts = append([]TaskOptions{TaskHooks(defaultTaskHooks...)}, opts...)
	return newTask(taskFunc, opts...)
}

// RunOnce schedules a task that run only once at specified time
// Note: any options affecting repeat rate (AtRate, WithDelay, etc.) would take no effect
func RunOnce(taskFunc TaskFunc, opts ...TaskOptions) (TaskCanceller, error) {
	opts = append([]TaskOptions{TaskHooks(defaultTaskHooks...)}, opts...)
	opts = append(opts, runOnceOption())
	return newTask(taskFunc, opts...)
}

func AddDefaultHook(hooks ...TaskHook) {
	defaultTaskHooks = append(defaultTaskHooks, hooks...)
	order.SortStable(defaultTaskHooks, order.OrderedFirstCompare)
}

/**************************
	Options
 **************************/

// Name option to give the task a name
func Name(name string) TaskOptions {
	return func(opt *TaskOption) error {
		opt.name = name
		return nil
	}
}

// TaskHooks option to add TaskHook
func TaskHooks(hooks ...TaskHook) TaskOptions {
	return func(opt *TaskOption) error {
		opt.hooks = append(opt.hooks, hooks...)
		return nil
	}
}

// StartAt option to set task's initial trigger time, should be future time
// Exclusive with StartAfter
func StartAt(startTime time.Time) TaskOptions {
	return func(opt *TaskOption) error {
		opt.initialTime = startTime
		return nil
	}
}

// StartAfter option to set task's initial trigger delay, should be positive duration
// Exclusive with StartAt
func StartAfter(delay time.Duration) TaskOptions {
	return func(opt *TaskOption) error {
		if delay < 0 {
			return fmt.Errorf("StartAfter doesn't support negative value")
		}
		return StartAt(time.Now().Add(delay))(opt)
	}
}

// AtRate option for "Fixed Interval" mode. Triggered every given interval.
// Long-running tasks overlap each other.
// Exclusive with WithDelay
func AtRate(repeatInterval time.Duration) TaskOptions {
	return func(opt *TaskOption) error {
		opt.mode = ModeFixedRate
		opt.interval = repeatInterval
		return nil
	}
}

// WithDelay option for "Fixed Delay" mode. Triggered with given delay after previous task finished
// Long-running tasks will never overlap
// Exclusive with AtRate
func WithDelay(repeatDelay time.Duration) TaskOptions {
	return func(opt *TaskOption) error {
		opt.mode = ModeFixedDelay
		opt.interval = repeatDelay
		return nil
	}
}

// CancelOnError option that automatically cancel the scheduled task if any execution returns non-nil error
func CancelOnError() TaskOptions {
	return func(opt *TaskOption) error {
		opt.cancelOnError = true
		return nil
	}
}

/**************************
	Helpers
 **************************/

func runOnceOption() TaskOptions {
	return func(opt *TaskOption) error {
		opt.mode = ModeRunOnce
		return nil
	}
}

func dynamicNext(nextFn nextFunc) TaskOptions {
	return func(opt *TaskOption) error {
		opt.mode = ModeDynamic
		opt.nextFunc = nextFn
		return nil
	}
}

