package dlock

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/xsync"
	"fmt"
	"github.com/hashicorp/consul/api"
	"sync"
	"time"
)

const (
	// lockFlagValue is a magic flag we set to indicate a key is being used for a lock.
	// It is used to detect a potential conflict with a semaphore.
	lockFlagValue = 0x275f2b610e0c3019
)

var (
	ErrCancelled       = fmt.Errorf("cancelled or timed out")
	ErrFailedToAcquire = fmt.Errorf("cannot aquire lock")
)

type ConsulLockOptions func(opt *ConsulLockOption)
type ConsulLockOption struct {
	SessionFunc      func(context.Context) (string, error)
	Key              string // Must be set and have write permissions
	Value            []byte // Optional, value to associate with the lock
	SessionTTL       time.Duration
	LockDelay        time.Duration
	MonitorRetries   int           // Optional, defaults to 0 which means no retries
	MonitorRetryTime time.Duration // Optional, defaults to 2 seconds
	QueryWaitTime    time.Duration // how long we block for at a time to check if lock acquisition is possible. This affects the minimum time it takes to cancel
}

type consulLockState int

const (
	stateUnknown consulLockState = iota
	stateAcquiring
	stateAcquired
	stateError
)

// ConsulLock implements Lock interface using consul lock described at https://www.consul.io/docs/guides/leader-election.html
// The implementation is modified api.Lock. The major difference are:
// - Session is created/maintained outside. There is no session creation when attempt to lock
// - "lock or wait" vs "try lock and return" is not pre-determined via options.
type ConsulLock struct {
	mtx    sync.Mutex
	client *api.Client
	option ConsulLockOption
	// State Variables, requires mutex lock to read and write
	lockLostCh     chan struct{}
	loopContext    context.Context
	loopCancelFunc context.CancelFunc
	state          consulLockState
	stateCond      *xsync.Cond
	session        string
	refreshFunc    context.CancelFunc // used when current acquisition should be stopped and restarted
	lastErr        error
}

func newConsulLock(client *api.Client, opts ...ConsulLockOptions) *ConsulLock {
	ret := ConsulLock{
		client: client,
		option: ConsulLockOption{
			MonitorRetries:   0,
			MonitorRetryTime: 2 * time.Second,
			QueryWaitTime:    15 * time.Second,
		},
	}
	ret.stateCond = xsync.NewCond(&ret.mtx)

	for _, fn := range opts {
		fn(&ret.option)
	}
	return &ret
}

func (l *ConsulLock) Key() string {
	return l.option.Key
}

// Lock attempts to acquire the lock and blocks while doing so.
// Providing a cancellable context.Context can be used to abort the lock attempt.
//
// Returns a channel that is closed if our lock is lost or an error.
// This channel could be closed at any time due to session invalidation,
// communication errors, operator intervention, etc. It is NOT safe to
// assume that the lock is held until Unlock() unless the Session is specifically
// created without any associated health checks. By default Consul sessions
// prefer liveness over safety and an application must be able to handle
// the lock being lost.
func (l *ConsulLock) Lock(ctx context.Context) error {
	// TODO
	l.start()
	if e := l.waitForAcquisition(ctx); e != nil {
		return e
	}
	return nil
}

func (l *ConsulLock) TryLock(ctx context.Context) error {
	// TODO
	return nil
}

func (l *ConsulLock) Release() error {
	// TODO
	return l.release()
}

func (l *ConsulLock) Lost() <-chan struct{} {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return l.lockLostCh
}

func (l *ConsulLock) start() {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	// Check if we're already maintaining the lock
	if l.loopContext == nil {
		l.startLoop()
	}
	return
}

func (l *ConsulLock) waitForAcquisition(ctx context.Context) error {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	for {
		switch l.state {
		case stateAcquired:
			return nil
		case stateError:
			return l.lastErr
		default:
			switch e := l.stateCond.Wait(ctx); e {
			case context.Canceled, context.DeadlineExceeded:
				return ErrFailedToAcquire
			}
		}
	}
}

func (l *ConsulLock) updateState(s consulLockState, setters ...func()) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	defer func(from, to consulLockState) {
		l.stateCond.Broadcast()
	}(l.state, s)

	for _, fn := range setters {
		fn()
	}
	l.state = s
}

func (l *ConsulLock) lastError() error {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return l.lastErr
}

// startLoop kickoff lock loop. mutex lock is required when call this function
func (l *ConsulLock) startLoop() {
	l.loopContext, l.loopCancelFunc = context.WithCancel(context.Background())
	if l.lockLostCh != nil {
		close(l.lockLostCh)
	}
	l.lockLostCh = make(chan struct{}, 1)
	go l.lockLoop(l.loopContext, l.loopCancelFunc)
}

// stopLoop stop lock loop. mutex lock is required when call this function
func (l *ConsulLock) stopLoop() {
	l.loopCancelFunc()
	if l.lockLostCh != nil {
		close(l.lockLostCh)
	}
	l.lockLostCh = nil
	l.loopContext = nil
	l.loopCancelFunc = nil
}

// refresh is called by session manager to notify potential change of session ID
func (l *ConsulLock) refresh() {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	if l.refreshFunc != nil {
		l.refreshFunc()
	}
}

// lockLoop is the main loop of attempting to maintain the lock.
// The lock state loop between Acquiring, Acquired and Error
// When unable to maintain the lock, the loop cancel the current context and try to start a new one
// Note: given context may also be cancelled outside, e.g. lock is released
func (l *ConsulLock) lockLoop(ctx context.Context, cancelFunc context.CancelFunc) {
	defer cancelFunc()
LOOP:
	for {
		select {
		case <-ctx.Done():
			l.updateState(stateUnknown)
			return
		default:
		}

		// update current state
		refreshCtx, fn := context.WithCancel(ctx)
		l.updateState(stateAcquiring, func() {
			l.refreshFunc = fn
		})

		// grab current session.
		// Note: in case of error, we don't reset previously used session,
		// 		 the release function will try to release lock using previously used session
		session, e := l.option.SessionFunc(ctx)
		switch {
		case e == context.Canceled || e == context.DeadlineExceeded:
			// current acquisition is cancelled
			continue
		case e != nil:
			l.updateState(stateAcquiring, func() {
				l.lastErr = fmt.Errorf("session is not available: %v", e)
			})
			continue
		default:
			l.updateState(stateAcquiring, func() {
				l.session = session
			})
		}

		// try to acquire lock
		switch e := l.acquireLock(refreshCtx, session, 0); e {
		case context.Canceled, context.DeadlineExceeded:
			// current acquisition is cancelled
			continue
		case nil:
			// lock acquired, continue
			l.updateState(stateAcquired)
			logger.WithContext(refreshCtx).Debugf("acquired lock [%s]", l.option.Key)
		default:
			l.updateState(stateError, func() {
				l.lastErr = e
			})
			continue
		}

		// up to this point, we have acquired the lock. enter monitor state
		// TODO use real TTL
		switch e := l.monitorLock(refreshCtx, session, 10*time.Second); e {
		case context.Canceled, context.DeadlineExceeded:
			// current acquisition is cancelled
			continue
		default:
			// we lost the lock
			logger.WithContext(refreshCtx).Debugf("lost lock [%s] - %v", l.option.Key, e)
			break LOOP
		}
	}

	// we lost lock, try to restart loop to compete for re-acquiring
	l.updateState(stateUnknown, func() {
		// the lock might be manually released before we lock the mutex, check if we still need to restart loop
		if l.loopContext == nil {
			return
		}
		l.startLoop()
	})
}

func (l *ConsulLock) acquireLock(ctx context.Context, session string, maxWait time.Duration) error {
	// Try to acquire lock
	kv := l.client.KV()
	pair := &api.KVPair{
		Key:     l.option.Key,
		Value:   l.option.Value,
		Session: session,
		Flags:   lockFlagValue,
	}

	waitUntilAvailable := maxWait <= 0
	var waitCtx context.Context
	if waitUntilAvailable {
		waitCtx = ctx
	} else {
		var cancelFunc context.CancelFunc
		waitCtx, cancelFunc = context.WithTimeout(ctx, maxWait)
		defer cancelFunc()
	}

LOOP:
	for {
		// try acquire
		switch acquired, _, e := kv.Acquire(pair, nil); {
		case e != nil:
			return fmt.Errorf("failed to acquire lock: %v", e)
		case acquired:
			break LOOP
		}

		// handle failure, might wait until lock become available and try again
		switch current, e := l.handleAcquisitionFailure(waitCtx, session, waitUntilAvailable); {
		case e != nil:
			return e
		case current == session:
			break LOOP
		case current != "" && !waitUntilAvailable:
			return ErrFailedToAcquire
		}

		// pause and retry
		// TODO use option
		select {
		case <-time.After(1 * time.Second):
		case <-waitCtx.Done():
			return ErrCancelled
		}
	}

	// up to this point, we acquired the lock
	return nil
}

// handleAcquisitionFailure handles lock acquisition failure. The provided ctx must be a cancellable context
// The function blocks until one of following condition is meet:
//
// 1. the provided context is cancelled or timed out
// 2. When waitUntilAvailable == true:
//    2.1 the lock becomes available (lock is not held any session)
//    2.2 the lock is held by its own session (this normally shouldn't happen)
// 3. When waitUntilAvailable == false:
//    3.1 current state of the lock become available (regardless if lock is available)
//
// Note 1: this function will block if consul is not available, unless context is cancelled
// Note 2: when this function returns, the lock might be in lock-delay period, meaning no session can acquire lock.
func (l *ConsulLock) handleAcquisitionFailure(ctx context.Context, session string, waitUntilAvailable bool) (currentOwner string, err error) {
	kv := l.client.KV()
	qOpts := &api.QueryOptions{}
	qOpts = qOpts.WithContext(ctx)

	for i := 0; true; i++ {
		logger.WithContext(ctx).Infof("wait attempt %d, WaitIndex=%d", i, qOpts.WaitIndex)
		// Look for an existing lock, blocking until not taken
		pair, meta, e := kv.Get(l.option.Key, qOpts)
		var owner string
		switch {
		case e != nil && !api.IsRetryableError(e):
			return "", fmt.Errorf("failed to read lock: %v", e)
		case pair != nil && pair.Flags != lockFlagValue:
			return "", api.ErrLockConflict
		case pair != nil:
			owner = pair.Session
		}

		// potentially retryable situations
		switch {
		case e != nil:
			// when error is retryable, we let loop continue
			logger.WithContext(ctx).Debugf("retryable error: %v", e)
		case !waitUntilAvailable:
			return owner, nil
		default:
			switch owner {
			case session, "":
				// the lock is held by current session OR the lock is not held by any session
				return owner, nil
			}
		}

		// see if cancelled
		select {
		case <-ctx.Done():
			return owner, ErrCancelled
		default:
		}

		// up to this point, we know the lock is held by other session, and context is not cancelled or timed out,
		qOpts.WaitIndex = meta.LastIndex
	}
	return
}

// monitorLock is a long running routine to monitor a lock ownership
// the function returns when given session lost ownership, or given session might not be valid
func (l *ConsulLock) monitorLock(ctx context.Context, session string, sessionTTL time.Duration) error {
	// TODO
	kv := l.client.KV()
	opts := &api.QueryOptions{
		RequireConsistent: true,
		WaitTime:          sessionTTL / 2,
	}
	opts = opts.WithContext(ctx)

	var err error
LOOP:
	for lastCheck := time.Now(); time.Since(lastCheck) < sessionTTL; {
		pair, meta, e := kv.Get(l.option.Key, opts)
		switch {
		case e != nil && api.IsRetryableError(e):
			opts.WaitIndex = 0
			err = e
		case e == nil && pair != nil && pair.Session == session:
			opts.WaitIndex = meta.LastIndex
			lastCheck = time.Now()
			err = nil
		case e == nil:
			err = fmt.Errorf("lock revoked by server")
			break LOOP
		default:
			err = e
			break LOOP
		}
	}
	select {
	case <-ctx.Done():
		return context.Canceled
	default:
		return err
	}
}

func (l *ConsulLock) release() error {
	// Hold the lock as we try to release
	l.mtx.Lock()
	defer l.mtx.Unlock()

	// Ensure the lock is active
	if l.loopContext == nil {
		return nil
	}

	// Stop lock loop
	l.stopLoop()

	// Release the lock explicitly if previously used session is known
	if l.session == "" {
		return nil
	}
	pair := &api.KVPair{
		Key:     l.option.Key,
		Session: l.session,
		Flags:   lockFlagValue,
	}

	_, _, err := l.client.KV().Release(pair, nil)
	if err != nil {
		return fmt.Errorf("failed to release lock: %v", err)
	}

	return nil
}
