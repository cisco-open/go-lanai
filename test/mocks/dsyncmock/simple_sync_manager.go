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

package dsyncmock

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/dsync"
	"go.uber.org/fx"
	"sync"
)

type SimpleSyncManagerMock struct {}

type NoopOut struct {
	fx.Out
	TestSyncManager dsync.SyncManager `group:"dsync"`
}

func ProvideNoopSyncManager() NoopOut {
	return NoopOut{
		TestSyncManager: SimpleSyncManagerMock{},
	}
}

func (m SimpleSyncManagerMock) Lock(key string, _ ...dsync.LockOptions) (dsync.Lock, error) {
	return &AlwaysLockMock{key: key}, nil
}

type AlwaysLockMock struct {
	mtx sync.Mutex
	key string
	ch  chan struct{}
}

func (l *AlwaysLockMock) Key() string {
	return l.key
}

func (l *AlwaysLockMock) Lock(_ context.Context) error {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	if l.ch == nil {
		l.ch = make(chan struct{}, 1)
	}
	return nil
}

func (l *AlwaysLockMock) TryLock(ctx context.Context) error {
	return l.Lock(ctx)
}

func (l *AlwaysLockMock) Release() error {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	if l.ch != nil {
		close(l.ch)
		l.ch = nil
	}
	return nil
}

func (l *AlwaysLockMock) Lost() <-chan struct{} {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return l.ch
}





