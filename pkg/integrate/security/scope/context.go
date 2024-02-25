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

package scope

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "time"
)

const (
	FxGroup = "security-scope"
)

var (
	scopeManager ScopeManager
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
	tenantExternalId string // target tenantExternalId
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

func (s Scope) String() string {
	user := s.userId
	if s.username != "" {
		user = s.username
	}
	tenant := s.tenantExternalId
	if s.tenantId != "" {
		tenant = s.tenantId
	}
	if tenant == "" {
		return user
	}
	return fmt.Sprintf("%s@%s", user, tenant)
}

func (s *Scope) Do(ctx context.Context, fn func(ctx context.Context)) (err error) {
	c, e := s.start(ctx)
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
	scopeManager.End(c)
	return nil
}

func (s *Scope) start(ctx context.Context) (context.Context, error) {
	if scopeManager == nil {
		return nil, ErrNotInitialized
	}
	return  scopeManager.StartScope(ctx, s)
}

func (s *Scope) validate(_ context.Context) error {
	if s.username != "" && s.userId != "" {
		return ErrUserIdAndUsernameExclusive
	}
	if s.tenantExternalId != "" && s.tenantId != "" {
		return ErrTenantIdAndNameExclusive
	}
	return nil
}

type ScopeManager interface {
	StartScope(ctx context.Context, scope *Scope) (context.Context, error)
	Start(ctx context.Context, opts...Options) (context.Context, error)
	End(ctx context.Context) context.Context
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

func Describe(ctx context.Context) string {
	scope, ok := ctx.Value(ctxKeyScope).(*Scope)
	if !ok {
		return "no scope"
	}
	return scope.String()
}

/**************************
	TestHooks
 **************************/

//goland:noinspection GoNameStartsWithPackageName
type ScopeOperationHook func(ctx context.Context, scope *Scope) context.Context

type ManagerCustomizer interface {
	Customize() []ManagerOptions
}

type ManagerCustomizerFunc func() []ManagerOptions
func (fn ManagerCustomizerFunc) Customize() []ManagerOptions {
	return fn()
}

func BeforeStartHook(hook ScopeOperationHook) ManagerOptions {
	return func(opt *managerOption) {
		opt.BeforeStartHooks = append(opt.BeforeStartHooks, hook)
	}
}


func AfterEndHook(hook ScopeOperationHook) ManagerOptions {
	return func(opt *managerOption) {
		opt.AfterEndHooks = append(opt.AfterEndHooks, hook)
	}
}

/**************************
	Context
 **************************/

type rollbackCtxKey struct{}
type scopeCtxKey struct{}

var ctxKeyRollback = rollbackCtxKey{}
var ctxKeyScope = scopeCtxKey{}

// scopedContext helps managerBase to backtrace context used for managerBase.DoStartScope and keep track of Scope
func newScopedContext(parent context.Context, scope *Scope, auth security.Authentication) context.Context {
	scoped := utils.NewMutableContext(parent, func(key interface{}) interface{} {
		switch key {
		case ctxKeyRollback:
			return parent
		case ctxKeyScope:
			return scope
		default: return nil
		}
	})
	security.MustSet(scoped, auth)
	return scoped
}

