package redisdsync

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cisco-open/go-lanai/pkg/dsync"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/pkg/utils/xsync"
	"github.com/go-redsync/redsync/v4"
	"sync"
	"time"
)

type lockState int

const (
	stateUnknown lockState = iota
	stateAcquired
	stateError
)

type RedisLockOptions func(opt *RedisLockOption)
type RedisLockOption struct {
	Context context.Context
	Name    string
	Valuer  dsync.LockValuer
	// AutoExpiry how long the acquired lock expires (released) in case the application crashes
	AutoExpiry time.Duration
	// RetryDelay how long we wait after a retryable error (usually network error)
	RetryDelay time.Duration
	// TimeoutFactor used to calculate redis CMD timeout when acquiring, extending and releasing lock.
	// timeout = AutoExpiry * TimeoutFactor
	TimeoutFactor float64
	// MaxExtendRetries how many times we attempt to extend the lock before give up
	MaxExtendRetries int
}

type RedisLock struct {
	mtx     sync.Mutex
	rsMutex *redsync.Mutex
	option  RedisLockOption
	// State Variables, requires mutex lock to read and write
	loopContext    context.Context
	loopCancelFunc context.CancelFunc
	lockLostCh     chan struct{}
	state          lockState
	stateCond      *xsync.Cond
	lastErr        error
}

func newRedisLock(rs *redsync.Redsync, opts ...RedisLockOptions) (lock *RedisLock) {
	opt := RedisLockOption{
		Valuer: dsync.NewJsonLockValuer(map[string]string{
			"name": "redis distributed lock",
		}),
		AutoExpiry:       10 * time.Second,
		RetryDelay:       500 * time.Millisecond,
		MaxExtendRetries: 3,
		TimeoutFactor:    0.05,
	}
	for _, fn := range opts {
		fn(&opt)
	}
	// Note: we only use TryLock and perform indefinite retries, so WithTries is set to 1 in order get proper error
	// See redsync.Mutex.TryLockContext for details
	rsMutex := rs.NewMutex(opt.Name,
		redsync.WithExpiry(opt.AutoExpiry),
		redsync.WithTries(1),
		redsync.WithTimeoutFactor(opt.TimeoutFactor),
		redsync.WithShufflePools(true),
		redsync.WithGenValueFunc(genValueFunc(opt.Valuer)),
	)

	// we start with a closed lost channel
	defer func() {
		lock.lockLostCh = make(chan struct{}, 1)
		close(lock.lockLostCh)
		lock.stateCond = xsync.NewCond(&lock.mtx)
	}()
	return &RedisLock{
		rsMutex: rsMutex,
		option:  opt,
	}
}

func (l *RedisLock) Key() string {
	return l.rsMutex.Name()
}

func (l *RedisLock) Lock(ctx context.Context) error {
	l.lazyStart()
	return l.waitForState(ctx, func(state lockState) (bool, error) {
		switch {
		case l.state == stateAcquired:
			return true, nil
		case l.loopContext == nil:
			return true, context.Canceled
		}
		return false, nil
	})
}

func (l *RedisLock) TryLock(ctx context.Context) error {
	l.lazyStart()
	// TryLock differ from Lock that it also return on any error state
	return l.waitForState(ctx, func(state lockState) (bool, error) {
		switch {
		case l.state == stateAcquired:
			return true, nil
		case l.state == stateError:
			return true, l.lastErr
		case l.loopContext == nil:
			return true, context.Canceled
		}
		return false, nil
	})
}

func (l *RedisLock) Release() error {
	return l.release()
}

func (l *RedisLock) Lost() <-chan struct{} {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return l.lockLostCh
}

func (l *RedisLock) lazyStart() {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	// Check if we're already maintaining the lock loop
	if l.loopContext == nil {
		l.startLoop()
	}
	return
}

func (l *RedisLock) waitForState(ctx context.Context, stateMatcher func(lockState) (bool, error)) error {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	for {
		if ok, e := stateMatcher(l.state); ok {
			return e
		}
		switch e := l.stateCond.Wait(ctx); {
		case errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded):
			return e
		}
	}
}

// updateState atomically update state, execute additional setters and broadcast the change.
// if given state < 0, only setters are executed
func (l *RedisLock) updateState(s lockState, setters ...func()) {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	for _, fn := range setters {
		fn()
	}

	if s < 0 {
		return
	}
	if s == stateAcquired && l.state != s {
		l.lockLostCh = make(chan struct{}, 1)
	} else if l.state == stateAcquired && l.state != s {
		close(l.lockLostCh)
	}

	if s == stateError || l.state != s {
		defer l.stateCond.Broadcast()
	}
	l.state = s
}

// startLoop kickoff lock loop. mutex lock is required when call this function
func (l *RedisLock) startLoop() {
	l.loopContext, l.loopCancelFunc = context.WithCancel(l.option.Context)
	go l.lockLoop(l.loopContext, l.loopCancelFunc)
}

// stopLoop stop lock loop. mutex lock is required when call this function
func (l *RedisLock) stopLoop() {
	if l.loopCancelFunc != nil {
		l.loopCancelFunc()
	}
	l.loopContext = nil
	l.loopCancelFunc = nil
}

// lockLoop is the main loop of attempting to maintain the lock.
// The lock state loop between Acquired and Error
// When unable to maintain the lock, the loop cancel the current context and try to lazyStart a new one
// Note: given context may also be cancelled outside, e.g. lock is released
func (l *RedisLock) lockLoop(ctx context.Context, cancelFunc context.CancelFunc) {
	defer cancelFunc()
	defer func() {
		// we've quited the loop, need some cleaning up:
		// 1. in case the lock is still locked (e.g. context canceled after lock is acquired), we need to explicitly release lock.
		_, _ = l.rsMutex.Unlock()
		l.updateState(stateUnknown)
	}()

LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		default:
		}

		// try to acquire lock
		switch e := l.rsMutex.TryLockContext(ctx); {
		case errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded):
			// current acquisition is cancelled
			continue
		case e == nil:
			// lock acquired, continue
			logger.WithContext(ctx).Debugf("acquired lock [%s]", l.option.Name)
			l.updateState(stateAcquired, func() { l.lastErr = nil })
		default:
			l.updateState(stateError, func() {
				l.lastErr = dsync.ErrLockUnavailable.WithMessage(`lock [%s] is held by another session`, l.option.Name).WithCause(e)
			})
			l.delay(ctx, l.option.RetryDelay)
			continue
		}

		// up to this point, we have acquired the lock. enter monitor state
		switch e := l.monitorLock(ctx); {
		case errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded):
			// current acquisition is cancelled
			continue
		default:
			// we lost the lock
			logger.WithContext(ctx).Debugf("lost lock [%s] - %v", l.option.Name, e)
			l.updateState(stateError, func() { l.lastErr = e })
		}
	}
}

// monitorLock is a long-running routine to monitor a lock ownership and try to extend lock periodically
// the function returns when given context is cancelled or lock lost ownership
func (l *RedisLock) monitorLock(ctx context.Context) error {
	var err error
	var failedAttempts int
	timeout := time.Duration(float64(l.option.AutoExpiry) * l.option.TimeoutFactor)
LOOP:
	for {
		var waitForExpiry bool
		expire := time.Until(l.rsMutex.Until())
		wait := expire / 2
		// Check if we have enough time to extend it. If not, we enter "wait for expiry" mode
		if wait < timeout || failedAttempts >= l.option.MaxExtendRetries {
			waitForExpiry = true
			wait = expire
		}
		if ok := l.delay(ctx, wait); !ok || waitForExpiry {
			break LOOP
		}

		// regardless the result, lock is not lost yet, if we cannot extend it now, we will try it later
		ok, e := l.rsMutex.ExtendContext(ctx)
		switch err = e; {
		case e == nil && !ok:
			err = dsync.ErrLockUnavailable.WithMessage(`failed to extend lock with unknown reason`)
		}
		if err != nil {
			failedAttempts++
			logger.WithContext(ctx).Debugf(e.Error())
		}
	}
	if err == nil {
		return context.Canceled
	}
	return err
}

// wait for given delay, return true if the delay is fulfilled (not cancelled by context)
func (l *RedisLock) delay(ctx context.Context, delay time.Duration) (success bool) {
	select {
	case <-time.After(delay):
		return true
	case <-ctx.Done():
		return false
	}
}

func (l *RedisLock) release() error {
	// Hold the lock as we try to release
	l.mtx.Lock()
	defer l.mtx.Unlock()

	// Ensure the lock is active
	if l.loopContext == nil {
		return nil
	}

	// Stop lock loop. Releasing the lock happens in the loop
	l.stopLoop()
	return nil
}

/***********************
	helpers
 ***********************/

type lockValue struct {
	Metadata interface{} `json:"metadata"`
	Token    string      `json:"token"`
}

func genValueFunc(valuer dsync.LockValuer) func() (string, error) {
	// attempt to parse metadata as JSON
	var value []byte
	if valuer != nil {
		value = valuer()
	}
	var meta interface{}
	if e := json.Unmarshal(value, &meta); e != nil {
		meta = value
	}
	v := lockValue{
		Metadata: meta,
		Token:    utils.RandomString(16),
	}
	data, e := json.Marshal(v)
	if e != nil {
		data = []byte(v.Token)
	}
	return func() (string, error) {
		return string(data), nil
	}
}
