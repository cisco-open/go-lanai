package dsyncmock

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/dsync"
	"go.uber.org/fx"
	"sync"
)

type SimpleSyncManagerMock struct {}

type NoopOut struct {
	fx.Out
	TestSyncManager dsync.SyncManager `group:"test"`
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





