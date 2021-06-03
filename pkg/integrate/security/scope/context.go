package scope

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
	"time"
)

var (
	scopeManager *defaultScopeManager
)

var (
	ErrNotInitialized             = fmt.Errorf("security scope manager is not initialied yet")
	ErrMissingDefaultSysAccount   = fmt.Errorf("unable to switch security scope: default system account is not configured")
	ErrMissingUser                = fmt.Errorf("unable to switch security scope: either username or user ID is required when not using default system account")
	ErrNotCurrentlyAuthenticated  = fmt.Errorf("unable to switch security scope without system account: current context is not authenticated")
	ErrUserIdAndUsernameExclusive = fmt.Errorf("invalid security scope option: username and user ID are exclusive")
	ErrTenantIdAndNameExclusive   = fmt.Errorf("invalid security scope option: tenant name and tenant ID are exclusive")
)

type Options func(*Scope)

type Scope struct {
	username   string // target username
	userId     string // target userId
	tenantName string // target tenantName
	tenantId   string // target tenantId
	time       time.Time
	useSysAcct bool
	cacheKey   *cKey
}

func New(opts ...Options) *Scope {
	scope := Scope{
		time: time.Now(),
	}
	for _, fn := range opts {
		fn(&scope)
	}
	return &scope
}

func (s *Scope) Start(ctx context.Context) (context.Context, error) {
	if scopeManager == nil {
		return nil, ErrNotInitialized
	}
	return  scopeManager.StartScope(ctx, s)
}

func (s *Scope) Do(ctx context.Context, fn func(ctx context.Context)) (err error) {
	if scopeManager == nil {
		return ErrNotInitialized
	}
	c, e := scopeManager.StartScope(ctx, s)
	if e != nil {
		return e
	}

	defer func() {
		switch e := recover().(type) {
		case nil:
		case error:
			err = e
		default:
			err = fmt.Errorf("%v", e)
		}
	}()

	fn(c)
	return nil
}

func (s *Scope) validate(_ context.Context) error {
	if s.username != "" && s.userId != "" {
		return ErrUserIdAndUsernameExclusive
	}
	if s.tenantName != "" && s.tenantId != "" {
		return ErrTenantIdAndNameExclusive
	}
	return nil
}

/**************************
	Convenient Functions
 **************************/

// Do invoke given function in a security scope specified by Options
// e.g.:
// 	scope.Do(ctx, func(ctx context.Context) {
// 		// do something with ctx
// 	}, scope.WithUsername("a-user"), scope.UseSystemAccount())
func Do(ctx context.Context, fn func(ctx context.Context), opts ...Options) error {
	return New(opts...).Do(ctx, fn)
}


/**************************
	Context
 **************************/

type rollbackCtxKey struct{}
type scopeCtxKey struct{}

var ctxKeyRollback = rollbackCtxKey{}

// scopedContext helps managerBase to backtrace context used for managerBase.DoStartScope and keep track of Scope
type scopedContext struct {
	context.Context
	scope *Scope
	auth security.Authentication
}

func (c scopedContext) Value(key interface{}) interface{} {
	if key == security.ContextKeySecurity {
		return c.auth
	}

	switch key.(type) {
	case rollbackCtxKey:
		return c.Context
	case scopeCtxKey:
		return c.scope
	default:
		return c.Context.Value(key)
	}
}
