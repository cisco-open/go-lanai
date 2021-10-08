package scheduler

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

// task execute TaskFunc based on TaskOption. Also implements TaskCanceller
type task struct {
	mtx  sync.Mutex
	id     string
	task   TaskFunc
	option TaskOption
	cancel context.CancelFunc
	done chan error
	err  error
}

func newTask(taskFunc TaskFunc, opts ...TaskOptions) (TaskCanceller, error) {
	if taskFunc == nil {
		return nil, fmt.Errorf("task function cannot be nil")
	}

	id := uuid.New().String()
	t := task{
		id:   id,
		task: taskFunc,
		done: make(chan error, 1),
	}
	for _, fn := range opts {
		if e := fn(&t.option); e != nil {
			return nil, e
		}
	}

	if t.option.name != "" {
		t.id = fmt.Sprintf("%s-%s", t.option.name, id)
	}

	switch {
	case t.option.mode != ModeRunOnce && t.option.interval <= 0:
		return nil, fmt.Errorf("repeated task should have positive repeat interval")
	}

	// start and return
	t.start(context.Background())
	return &t, nil
}

// Cancel implements TaskCanceller
func (t *task) Cancel() {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.err = context.Canceled
	t.cancel()
}

// Cancelled implements TaskCanceller
func (t *task) Cancelled() <-chan error {
	return t.done
}

// start main loop
func (t *task) start(ctx context.Context) {
	taskCtx, fn := context.WithCancel(ctx)
	t.cancel = fn
	go t.loop(taskCtx)
}

// loop is the main loop for the task
func (t *task) loop(ctx context.Context) {
	defer func() {
		t.mtx.Lock()
		defer t.mtx.Unlock()
		t.done <- t.err
		close(t.done)
	}()

	// first, figure out first fire time if set
	var delay time.Duration
	if !t.option.initialTime.IsZero() {
		delay = time.Until(t.option.initialTime)
		if delay < 0 {
			if t.option.mode == ModeFixedRate {
				// adjust using interval (first positive trigger time)
				delay = (t.option.interval + (delay % t.option.interval)) % t.option.interval
			} else {
				delay = 0
			}
		}
	}
	select {
	case <-time.After(delay):
		t.execTask(ctx, t.option.mode != ModeFixedRate)
	case <-ctx.Done():
		return
	}

	// repeat if applicable
	switch t.option.mode {
	case ModeFixedRate:
		t.fixedIntervalLoop(ctx)
	case ModeFixedDelay:
		t.fixedDelayLoop(ctx)
	case ModeRunOnce:
	}
}

func (t *task) fixedIntervalLoop(ctx context.Context) {
	ticker := time.NewTicker(t.option.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			t.execTask(ctx, false)
		case <-ctx.Done():
			return
		}
	}
}

func (t *task) fixedDelayLoop(ctx context.Context) {
	for {
		select {
		case <-time.After(t.option.interval):
			t.execTask(ctx, true)
		case <-ctx.Done():
			return
		}
	}
}

func (t *task) execTask(ctx context.Context, wait bool) {
	errCh := make(chan error, 1)
	go func() {
		execCtx := ctx
		var err error
		defer func() {
			// try recover
			if e := recover(); e != nil {
				err = fmt.Errorf("%v", e)
			}

			// post-hook
			for _, hook := range t.option.hooks {
				hook.AfterTrigger(execCtx, t.id, err)
			}

			// handle error
			if err != nil {
				t.handleError(execCtx, err)
			}

			// notify and cleanup
			errCh <- err
			close(errCh)
		}()

		// pre-hook
		for _, hook := range t.option.hooks {
			execCtx = hook.BeforeTrigger(execCtx, t.id)
		}

		// run task
		err = t.task(execCtx)
	}()

	if !wait {
		return
	}

	select {
	case <-errCh:
	}
}

func (t *task) handleError(ctx context.Context, err error) {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	t.err = err
	if t.option.cancelOnError {
		logger.WithContext(ctx).Infof("Task [%s] cancelled due to error: %v", t.id, err)
		t.cancel()
	} else {
		logger.WithContext(ctx).Debugf("Task [%s] returned with error: %v", t.id, err)
	}
}
