package loop

import (
	"context"
	"time"
)

// TaskFunc can be scheduled to Loop with TaskOptions
type TaskFunc func(ctx context.Context, l *Loop) (ret interface{}, err error)

// RepeatIntervalFunc is used when schedule repeated TaskFunc.
// it takes result and error of previous TaskFunc invocation and determine the delay of next TaskFunc invocation
type RepeatIntervalFunc func(result interface{}, err error) time.Duration

type TaskOptions func(opt *TaskOption)
type TaskOption struct {
	RepeatIntervalFunc RepeatIntervalFunc
}

type task struct {
	f     TaskFunc
	errCh chan error
	retCh chan interface{}
	opt   TaskOption
}

type Loop struct {
	stopCh chan struct{}
	taskCh chan *task
}

func NewLoop() *Loop{
	return &Loop {
		stopCh: make(chan struct{}),
		taskCh: make(chan *task),
	}
}

func (l *Loop) Run(ctx context.Context) (context.Context, context.CancelFunc) {
	ctxWithCancel, cFunc := context.WithCancel(ctx)
	go l.loop(ctxWithCancel)
	return ctxWithCancel, cFunc
}

func (l *Loop) Repeat(tf TaskFunc, opts ...TaskOptions) {
	opt := TaskOption{
		RepeatIntervalFunc: fixedRepeatIntervalFunc(10 * time.Millisecond),
	}
	for _, f := range opts {
		f(&opt)
	}
	l.taskCh <- &task{
		f: l.makeTaskFuncWithRepeat(tf, opt.RepeatIntervalFunc),
		opt: opt,
	}
}

func (l *Loop) Do(tf TaskFunc, opts ...TaskOptions) {
	opt := TaskOption{}
	for _, f := range opts {
		f(&opt)
	}
	l.taskCh <- &task{
		f: tf,
		opt: opt,
	}
}

func (l *Loop) DoAndWait(tf TaskFunc, opts ...TaskOptions) (interface{}, error) {
	opt := TaskOption{}
	for _, f := range opts {
		f(&opt)
	}
	errCh := make(chan error)
	retCh := make(chan interface{})
	l.taskCh <- &task{
		f: tf,
		errCh: errCh,
		retCh: retCh,
		opt: opt,
	}
	select {
	case e := <-errCh:
		return nil, e
	case r := <-retCh:
		return r, nil
	}
}

func (l *Loop) loop(ctx context.Context) {
	for {
		select {
		case t := <-l.taskCh:
			l.do(ctx, t)
		case <-ctx.Done():
			return
		}
	}
}

func (l *Loop) do(ctx context.Context, t *task) {
	// we assume the cancel signal is propagated from parent
	done := make(chan struct{})
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	// we guarantee that either errCh or retCh is pushed with value, so we don't need to explicitly close those channels
	defer cancelFunc()

	go func() {
		defer func(){
			done <- struct{}{}
		}()

		if r, e := t.f(cancelCtx, l); t.errCh != nil && e != nil {
			t.errCh <- e
		} else if t.retCh != nil {
			t.retCh <- r
		}
	}()

	select {
	case <-done:
	case <-cancelCtx.Done():
		if t.errCh != nil {
			t.errCh <- cancelCtx.Err()
		}
	}
	return
}

// makeTaskFuncWithRepeat make a func that execute given TaskFunc and reschedule itself after given "interval"
func (l *Loop) makeTaskFuncWithRepeat(tf TaskFunc, intervalFunc RepeatIntervalFunc) TaskFunc {
	return func(ctx context.Context, l *Loop) (ret interface{}, err error) {
		// reschedule after delayed time
		defer func() {
			interval := intervalFunc(ret, err)
			l.repeatAfter(l.makeTaskFuncWithRepeat(tf, intervalFunc), interval)
		}()
		ret, err = tf(ctx, l)
		return
	}
}

func (l *Loop) repeatAfter(tf TaskFunc, interval time.Duration) {
	go func() {
		time.Sleep(interval)
		l.Do(tf)
	}()
}