package dsync

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/xsync"
	"errors"
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
	shutdown    bool
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

func NewConsulLockManager(ctx *bootstrap.ApplicationContext, conn *consul.Connection, opts ...ConsulSessionOptions) (ret *ConsulSyncManager) {
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

func (m *ConsulSyncManager) Start(_ context.Context) error {
	// do nothing, we support lazy start
	return nil
}

func (m *ConsulSyncManager) Stop(ctx context.Context) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	// enter shutdown mode
	m.shutdown = true
	// stopLoop session loop
	m.stopLoop()
	// release all existing locks
	for k, l := range m.locks {
		if e := l.Release(); e != nil {
			logger.WithContext(ctx).Warnf("Failed to release lock [%s]: %v", k, e)
		}
	}
	return nil
}

func (m *ConsulSyncManager) Lock(key string, opts ...LockOptions) (Lock, error) {
	if key == "" {
		return nil, fmt.Errorf(`cannot create distributed lock: key is required but missing`)
	}

	option := LockOption{
		Valuer: NewJsonLockValuer(map[string]string{
			"name": fmt.Sprintf("distributed lock - %s", m.appCtx.Name()),
		}),
	}
	for _, fn := range opts {
		fn(&option)
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()

	if m.shutdown {
		return nil, ErrSyncManagerStopped
	} else if lock, ok := m.locks[key]; ok {
		return lock, nil
	}

	m.locks[key] = newConsulLock(m.client, func(opt *ConsulLockOption) {
		opt.Context = m.appCtx
		opt.SessionFunc = m.waitForSession
		opt.Key = key
		opt.Valuer = option.Valuer
	})
	return m.locks[key], nil
}

// startLoop requires mutex lock
func (m *ConsulSyncManager) startLoop() error {
	if m.shutdown {
		return ErrSyncManagerStopped
	}
	if m.cancelFunc == nil {
		ctx, cf := context.WithCancel(m.appCtx)
		m.cancelFunc = cf
		go m.sessionLoop(ctx)
	}
	return nil
}

// stopLoop requires mutex lock
func (m *ConsulSyncManager) stopLoop() {
	if m.cancelFunc != nil {
		m.cancelFunc()
		m.cancelFunc = nil
	}
	return
}

// waitForSession returns current managed session. It blocks until session is available or given context is cancelled
func (m *ConsulSyncManager) waitForSession(ctx context.Context) (string, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	//nolint:contextcheck // startLoop starts background goroutine, it needs ApplicationContext to ensure the loop is not cancelled by given ctx
	if e := m.startLoop(); e != nil {
		return "", e
	}
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

func (m *ConsulSyncManager) updateSession(sid string) {
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
			logger.WithContext(ctx).Infof("sync manager stopped")
			return
		default:
		}

		// reset session
		m.updateSession("")
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
			select {
			case <-time.After(m.option.RetryDelay):
				continue
			case <-ctx.Done():
				continue
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
		// session until a doneChan is closed. This is meant to be used in a long-running
		// goroutine to ensure a session stays valid.
		wOpts := (*api.WriteOptions)(nil).WithContext(ctx)
		e := m.client.Session().RenewPeriodic(m.option.TTL.String(), session, wOpts, ctx.Done())
		switch {
		case e == nil:
			// just continue
		case errors.Is(e, api.ErrSessionExpired):
			logger.WithContext(ctx).Warnf("session expired")
			return e
		default:
			logger.WithContext(ctx).Warnf("session lost: %v", e)
			return e
		}
	}
}
