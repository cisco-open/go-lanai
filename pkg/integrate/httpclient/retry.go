package httpclient

import (
	"context"
	"time"
)

type Retryable func(ctx context.Context) (interface{}, error)

// RetryCallback retry control function.
// RetryCallback is executed each time when non-nil error is returned.
// "n": indicate the iteration number of attempts.
// "rs" and "err" indicate the result of the current attempt
// RetryCallback returns whether Retryable need to keep trying and optionally wait for "backoff" before next attempt
type RetryCallback func(n int, rs interface{}, err error) (shouldContinue bool, backoff time.Duration)

// Try keep trying to execute the Retryable until
// 1. No error is returned
// 2. Timeout reached
// 3. RetryCallback tells it to stop
// The Retryable is executed in separated goroutine, and the RetryCallback is invoked in current goroutine
// when non-nil error is returned.
// If the execution finished without any successful result, latest error is returned if available, otherwise context.Err()
func (r Retryable) Try(ctx context.Context, timeout time.Duration, cb RetryCallback) (interface{}, error) {
	if cb == nil {
		cb = func(_ int, _ interface{}, _ error) (bool, time.Duration) {
			return true, 0
		}
	}

	type result struct {
		value interface{}
		err   error
	}
	timoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var lastErr error
	for i := 1; ; i++ {
		rsCh := make(chan result, 1)
		go func() {
			var rs result
			rs.value, rs.err = r(timoutCtx)
			rsCh <- rs
			close(rsCh)
		}()

		select {
		case <-timoutCtx.Done():
			if lastErr == nil {
				lastErr = timoutCtx.Err()
			}
			return nil, lastErr
		case rs := <-rsCh:
			if rs.err == nil {
				return rs.value, nil
			}
			lastErr = rs.err
			switch again, backoff := cb(i, rs.value, rs.err); {
			case !again:
				return rs.value, rs.err
			case backoff > 0:
				time.Sleep(backoff)
			}
		}
	}
}
