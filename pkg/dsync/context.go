// Package dlock provides distributed acquireLock support of microservices and provide common usage patterns
// around distributed acquireLock, such as acquireLock-based service leader election and Once implementation similar to sync package.
package dsync

import "context"

type SyncManager interface {

}

// Lock distributed mutex acquireLock backed by external infrastructure service such as consul or redis
type Lock interface {
	Key() string
	Lock(ctx context.Context) error
	TryLock(ctx context.Context) error
	Release() error
	Lost() <-chan struct{}
}
