package dlock

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/xsync"
	"fmt"
	"github.com/hashicorp/consul/api"
	"sync"
	"time"
)

// ConsulSyncManager implements SyncManager leveraging consul's session feature
// See 	https://learn.hashicorp.com/tutorials/consul/application-leader-elections?in=consul/developer-configuration
//		https://learn.hashicorp.com/tutorials/consul/distributed-semaphore
// 		https://www.consul.io/docs/dynamic-app-config/sessions
type ConsulSyncManager struct {
	mtx         sync.Mutex
	appCtx      *bootstrap.ApplicationContext
	client      *api.Client
	option      ConsulSessionOption
	session     string
	sessionCond *xsync.Cond
	cancelFunc  context.CancelFunc
	locks       map[string]*ConsulLock
}

type ConsulSessionOptions func(opt *ConsulSessionOption)
type ConsulSessionOption struct {
	Name       string
	TTL        time.Duration
	LockDelay  time.Duration
	RetryDelay time.Duration
}

func newConsulLockManager(ctx *bootstrap.ApplicationContext, conn *consul.Connection, opts ...ConsulSessionOptions) (ret *ConsulSyncManager) {
	ret = &ConsulSyncManager{
		appCtx: ctx,
		client: conn.Client(),
		option: ConsulSessionOption{
			Name:       fmt.Sprintf("%s", ctx.Name()),
			TTL:        10 * time.Second,
			LockDelay:  2 * time.Second,
			RetryDelay: 2 * time.Second,
		},
		locks: make(map[string]*ConsulLock),
	}
	ret.sessionCond = xsync.NewCond(&ret.mtx)

	for _, fn := range opts {
		fn(&ret.option)
	}
	return
}

func (m *ConsulSyncManager) Lock(key string) (Lock, error) {
	if key == "" {
		return nil, fmt.Errorf(`cannot create distributed mutext: key is required but missing`)
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()

	if lock, ok := m.locks[key]; ok {
		return lock, nil
	}

	m.locks[key] = newConsulLock(m.client, func(opt *ConsulLockOption) {
		opt.SessionFunc = m.managedSessionFunc()
		opt.Key = key
		opt.Value = []byte("reserved value")
	})
	return m.locks[key], nil
}

func (m *ConsulSyncManager) start() {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if m.cancelFunc == nil {
		ctx, cf := context.WithCancel(m.appCtx)
		m.cancelFunc = cf
		go m.sessionLoop(ctx)
	}
	return
}

func (m *ConsulSyncManager) stop() {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if m.cancelFunc != nil {
		m.cancelFunc()
		m.cancelFunc = nil
	}
	return
}

// managedSessionFunc returns current session if a TODO
func (m *ConsulSyncManager) managedSessionFunc() func(context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		m.start()
		m.mtx.Lock()
		defer m.mtx.Unlock()
		for {
			switch m.session {
			case "":
				if e := m.sessionCond.Wait(ctx); e != nil {
					return "", e
				}
			default:
				return m.session, nil
			}
		}
	}
}

func (m *ConsulSyncManager) updateSession(sid string) {
	if sid == "" {
		return
	}
	m.mtx.Lock()
	defer m.mtx.Unlock()
	defer func(from, to string) {
		if from != to {
			m.sessionCond.Broadcast()
		}
	}(m.session, sid)
	m.session = sid
}

// sessionLoop is the main loop to manage session
func (m *ConsulSyncManager) sessionLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.Infof("sync manager stopped")
			return
		default:
		}

		// reset session
		m.updateSession("")
		// TODO use configuration
		wOpts := (*api.WriteOptions)(nil).WithContext(ctx)
		session, _, e := m.client.Session().Create(&api.SessionEntry{
			Name:      m.option.Name,
			TTL:       m.option.TTL.String(),
			LockDelay: m.option.LockDelay,
			Behavior:  "delete",
		}, wOpts)

		switch e {
		case nil:
			m.updateSession(session)
		default:
			if api.IsRetryableError(e) {
				select {
				case <-time.After(m.option.RetryDelay):
					continue
				case <-ctx.Done():
					continue
				}
			}
		}

		// keep renewing the session
		_ = m.keepSession(ctx, session)

		// session is invalid/expired by this point.
		// try to notify all existing locks
		m.mtx.Lock()
		for _, l := range m.locks {
			l.refresh()
		}
		m.mtx.Unlock()
	}
}

func (m *ConsulSyncManager) keepSession(ctx context.Context, session string) error {
	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		// RenewPeriodic is used to periodically invoke Session.Renew on a
		// session until a doneChan is closed. This is meant to be used in a long running
		// goroutine to ensure a session stays valid.
		wOpts := (*api.WriteOptions)(nil).WithContext(ctx)
		e := m.client.Session().RenewPeriodic(m.option.TTL.String(), session, wOpts, ctx.Done())
		switch e {
		case nil:
			// just continue
		case api.ErrSessionExpired:
			logger.Warnf("session expired")
			return e
		default:
			logger.Debugf("session renew error: %v", e)
			return e
		}
	}
}
