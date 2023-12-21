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

package xsync

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestCondRace(t *testing.T) {
	x := 0
	c := NewCond(&sync.Mutex{})
	wg := sync.WaitGroup{}
	wg.Add(3)
	errMtx := sync.Mutex{}
	var err error
	go func() {
		c.L.Lock()
		x = 1
		_ = c.Wait(context.Background())
		if x != 2 {
			errMtx.Lock()
			err = fmt.Errorf("x should be 2 after Wait")
			errMtx.Unlock()
		}
		x = 3
		c.Broadcast()
		c.L.Unlock()
		wg.Done()
	}()
	go func() {
		c.L.Lock()
		for {
			if x == 1 {
				x = 2
				c.Broadcast()
				break
			}
			c.L.Unlock()
			runtime.Gosched()
			c.L.Lock()
		}
		c.L.Unlock()
		wg.Done()
	}()
	go func() {
		c.L.Lock()
		for {
			if x == 2 {
				_ = c.Wait(context.Background())
				if x != 3 {
					errMtx.Lock()
					err = fmt.Errorf("x should be 3 after Wait")
					errMtx.Unlock()
				}
				break
			}
			if x == 3 {
				break
			}
			c.L.Unlock()
			runtime.Gosched()
			c.L.Lock()
		}
		c.L.Unlock()
		wg.Done()
	}()

	wg.Wait()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCondCancel(t *testing.T) {
	c := NewCond(&sync.Mutex{})
	testCtx, cancelFunc := context.WithTimeout(context.Background(), 500 * time.Millisecond)
	defer cancelFunc()
	err := fmt.Errorf("func Wait(context.Context) should finish when wait time exceeded")

	go func() {
		waitCtx, waitCancelFunc := context.WithTimeout(context.Background(), 200 * time.Millisecond)
		defer waitCancelFunc()
		c.L.Lock()
		e := c.Wait(waitCtx)
		if e != context.DeadlineExceeded {
			err = fmt.Errorf("func Wait(context.Context) should return context.DeadlineExceeded")
		} else {
			err = nil
		}
		c.L.Unlock()
		cancelFunc()
	}()
	select {
	case <-testCtx.Done():
	}
	if err != nil {
		t.Fatal(err)
	}
}
