package loop

import (
	"time"
)

// FixedRepeatInterval returns a TaskOptions which set repeat interval to be fixed duration
func FixedRepeatInterval(interval time.Duration) TaskOptions {
	return func(opt *TaskOption) {
		opt.RepeatIntervalFunc = fixedRepeatIntervalFunc(interval)
	}
}

func fixedRepeatIntervalFunc(interval time.Duration) RepeatIntervalFunc {
	return func(_ interface{}, _ error) time.Duration {
		return interval
	}
}

// ExponentialRepeatIntervalOnError returns a TaskOptions
// which set repeat interval to be exponentially increased if error is not nil.
// the repeat interval is reset to "init" if error is nil
func ExponentialRepeatIntervalOnError(init time.Duration, factor float64) TaskOptions {
	if factor < 1 {
		panic("attempt to use ExponentialRepeatIntervalOnError with a factor less than 1")
	}
	return func(opt *TaskOption) {
		opt.RepeatIntervalFunc = exponentialRepeatIntervalOnErrorFunc(init, factor)
	}
}

func exponentialRepeatIntervalOnErrorFunc(init time.Duration, factor float64) RepeatIntervalFunc {
	curr := init
	return func(_ interface{}, err error) time.Duration {
		if err == nil {
			curr = init
		} else {
			curr = time.Duration(float64(curr) * factor)
		}
		return curr
	}
}
