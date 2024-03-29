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
	"fmt"
	"sync"
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

type taskResult struct {
	ret interface{}
	err error
}

type task struct {
	f        TaskFunc
	resultCh chan taskResult
	opt      TaskOption
}

type Loop struct {
	taskCh   chan *task
	mtx      sync.Mutex
	ctx      context.Context
	cancelFn context.CancelFunc
}

func NewLoop() *Loop {
	return &Loop{
		taskCh: make(chan *task),
	}
}

func (l *Loop) Run(ctx context.Context) (context.Context, context.CancelFunc) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	if l.ctx == nil {
		l.ctx, l.cancelFn = context.WithCancel(ctx)
		go l.loop(l.ctx)
	}
	return l.ctx, l.cancelFn
}

func (l *Loop) Repeat(tf TaskFunc, opts ...TaskOptions) {
	opt := TaskOption{
		RepeatIntervalFunc: fixedRepeatIntervalFunc(10 * time.Millisecond),
	}
	for _, f := range opts {
		f(&opt)
	}
	l.taskCh <- &task{
		f:   l.makeTaskFuncWithRepeat(tf, opt.RepeatIntervalFunc),
		opt: opt,
	}
}

func (l *Loop) Do(tf TaskFunc, opts ...TaskOptions) {
	opt := TaskOption{}
	for _, f := range opts {
		f(&opt)
	}
	l.taskCh <- &task{
		f:   tf,
		opt: opt,
	}
}

func (l *Loop) DoAndWait(tf TaskFunc, opts ...TaskOptions) (interface{}, error) {
	opt := TaskOption{}
	for _, f := range opts {
		f(&opt)
	}
	resultCh := make(chan taskResult)
	defer close(resultCh)
	l.taskCh <- &task{
		f:        tf,
		resultCh: resultCh,
		opt:      opt,
	}
	select {
	case result := <-resultCh:
		return result.ret, result.err
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
	execCtx, doneFn := context.WithCancel(ctx)
	// we guarantee that either resultCh is pushed with value, so we don't need to explicitly close those channels here

	go func() {
		defer func() {
			if e := recover(); e != nil && t.resultCh != nil {
				t.resultCh <- taskResult{err: fmt.Errorf("%v", e)}
			}
			doneFn()
		}()

		r, e := t.f(execCtx, l)
		if t.resultCh != nil {
			// check if parent ctx is cancelled
			select {
			case <-ctx.Done():
				t.resultCh <- taskResult{err: ctx.Err()}
			default:
				t.resultCh <- taskResult{
					ret: r,
					err: e,
				}
			}
		}
	}()

	// wait for finish or cancelled
	select {
	case <-execCtx.Done():
	}
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
		timer := time.NewTimer(interval)
		select {
		case <-timer.C:
			l.Do(tf)
		case <-l.ctx.Done():
			timer.Stop()
		}
	}()
}
