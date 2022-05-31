package dsyncmock

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/dsync"
	"go.uber.org/fx"
)

type NoopSyncManagerMock struct {}

type NoopOut struct {
	fx.Out
	TestSyncManager dsync.SyncManager `group:"test"`
}

func ProvideNoopSyncManager() NoopOut {
	return NoopOut{
		TestSyncManager: NoopSyncManagerMock{},
	}
}

func (m NoopSyncManagerMock) Lock(key string, _ ...dsync.LockOptions) (dsync.Lock, error) {
	return NoopLockMock(key), nil
}

type NoopLockMock string

func (l NoopLockMock) Key() string {
	return string(l)
}

func (l NoopLockMock) Lock(_ context.Context) error {
	return nil
}

func (l NoopLockMock) TryLock(_ context.Context) error {
	return nil
}

func (l NoopLockMock) Release() error {
	return nil
}

func (l NoopLockMock) Lost() <-chan struct{} {
	// return a closed channel
	ch := make(chan struct{}, 1)
	defer close(ch)
	return ch
}





