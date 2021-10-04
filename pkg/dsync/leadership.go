package dsync

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"fmt"
	"sync"
)

const (
	leadershipLockKeyFormat = "service/%s/leadership"
)

var (
	leadershipOnce sync.Once
	leadershipLock Lock
)

// LeadershipLock returns globally maintained lock for leadership election
// To check leadership, use Lock.TryLock and check error.
// Example:
// 		if e := LeadershipLock().TryLock(ctx); e == nil {
// 			// do what a leader should do
//		}
//
// This function panic if it's call too soon during startup
// Note: Lock.Lost() channel should be monitored for long-running goroutine, since leadership could be revoked any time by operators
func LeadershipLock() Lock {
	if leadershipLock == nil {
		panic("Leadership Lock is not initialized yet")
	}
	return leadershipLock
}

func startLeadershipLock(ctx context.Context, di initDI) (err error) {
	leadershipOnce.Do(func() {
		leadershipLock, err = syncManager.Lock(
			fmt.Sprintf(leadershipLockKeyFormat, di.AppCtx.Name()),
			func(opt *LockOption) {
				opt.Valuer = leaderLockValuer(di.AppCtx)
			},
		)
		if err != nil {
			return
		}
		// Note we don't care the lock result, as long as we tell the lock to keep trying.
		// This goroutine is for logging purpose
		go func() {
		LOOP:
			for {
				if e := leadershipLock.Lock(di.AppCtx); e == nil {
					logger.WithContext(ctx).Infof("Leadership - become leader [%s]", leadershipLock.Key())
					select {
					case <-leadershipLock.Lost():
						logger.WithContext(ctx).Infof("Leadership - lost [%s]", leadershipLock.Key())
					case <-di.AppCtx.Done():
					}
				}
				select {
				case <-di.AppCtx.Done():
					break LOOP
				default:
				}
			}

		}()

	})
	return
}

func leaderLockValuer(appCtx *bootstrap.ApplicationContext) LockValuer {
	return NewJsonLockValuer(map[string]interface{}{
		"service": map[string]interface{}{
			"name":         appCtx.Name(),
			"port":         appCtx.Config().Value("server.port"),
			"context-path": appCtx.Config().Value("server.context-path"),
		},
		"build": bootstrap.BuildInfoMap,
	})
}
