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
	"sync"
	"sync/atomic"
	"unsafe"
)

// Cond is similar to sync.Cond: Conditional variable implementation that uses channels for notifications.
// This implementation differ from sync.Cond in following ways:
// - Only supports Broadcast
// - Wait with ctx that can be cancelled
// see sync.Cond for usage
// Ref: https://gist.github.com/zviadm/c234426882bfc8acba88f3503edaaa36#file-cond2-go
type Cond struct {
	L sync.Locker
	n unsafe.Pointer
}

func NewCond(l sync.Locker) *Cond {
	c := &Cond{L: l}
	n := make(chan struct{})
	c.n = unsafe.Pointer(&n)
	return c
}

// Wait for Broadcast calls. Similar to regular sync.Cond, this unlocks the underlying
// locker first, waits on changes and re-locks it before returning.
func (c *Cond) Wait(ctx context.Context) (err error) {
	n := c.notifyChan()
	c.L.Unlock()
	select {
	case <-n:
	case <-ctx.Done():
		err = ctx.Err()
	}
	c.L.Lock()
	return
}

// Broadcast call notifies everyone that something has changed.
func (c *Cond) Broadcast() {
	n := make(chan struct{})
	ptrOld := atomic.SwapPointer(&c.n, unsafe.Pointer(&n))
	close(*(*chan struct{})(ptrOld))
}

// notifyChan Returns a channel that can be used to wait for next Broadcast() call.
func (c *Cond) notifyChan() <-chan struct{} {
	ptr := atomic.LoadPointer(&c.n)
	return *((*chan struct{})(ptr))
}

