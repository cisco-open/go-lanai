package scope

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/seclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"errors"
	"fmt"
	"reflect"
	"time"
)

type authenticateFunc func(ctx context.Context, pKey *cKey) (security.Authentication, error)

type managerBase struct {
	cache              *cache
	tokenStoreReader   oauth2.TokenStoreReader
	failureBackOff     time.Duration
	guaranteedValidity time.Duration
	beforeStartHooks   []ScopeOperationHook
	afterEndHooks      []ScopeOperationHook
}

func (b *managerBase) DoStartScope(ctx context.Context, scope *Scope, authFunc authenticateFunc) (context.Context, error) {
	for _, hook := range b.beforeStartHooks {
		ctx = hook(ctx, scope)
	}

	auth, e := b.GetOrAuthenticate(ctx, scope.cacheKey, scope.time, authFunc)
	if e != nil {
		return nil, e
	}

	// set new security auth and return
	scoped := &scopedContext{
		Context: ctx,
		scope:   scope,
		auth:    auth,
	}
	return scoped, nil
}

func (b *managerBase) End(ctx context.Context) (ret context.Context) {
	rollback := ctx.Value(ctxKeyRollback)
	scope, _ := ctx.Value(ctxKeyScope).(*Scope)
	switch c := rollback.(type) {
	case context.Context:
		ret = c
	default:
		ret = ctx
	}
	for _, hook := range b.afterEndHooks {
		ret = hook(ret, scope)
	}
	return
}

func (b *managerBase) GetOrAuthenticate(ctx context.Context, pKey *cKey, rTime time.Time, authFunc authenticateFunc) (ret security.Authentication, err error) {
	return b.cache.GetOrLoad(ctx, pKey , b.cacheLoadFunc(rTime, authFunc), b.cacheValidateFunc())
}

func (b *managerBase) resolveUser(auth security.Authentication) (username, userId string, err error) {
	if !security.IsFullyAuthenticated(auth) {
		return "", "", fmt.Errorf("not currently authenticated")
	}

	switch details := auth.Details().(type) {
	case security.UserDetails:
		username = details.Username()
		userId = details.UserId()
	default:
		username, err = security.GetUsername(auth)
	}
	return
}

// normalizeTargetUser check if currently authenticated user is same user of target user
// if is same user, set target username and remove target userId
// use case:
// 		currently logged in as "user1" with userId="user1-id" and scope indicate target scope.userId="user1-id"
//		normalize result: scope.userId = "", scope.username="user1"
func (b *managerBase) normalizeTargetUser(auth security.Authentication, scope *Scope) {
	if scope.username == "" && scope.userId == "" || !b.isSameUser(scope.username, scope.userId, auth) {
		return
	}

	username, _, e := b.resolveUser(auth)
	if e != nil {
		return
	}

	scope.username = username
	scope.userId = ""
}

func (b *managerBase) prepareCacheKey(scope *Scope, srcUsername string) {
	scope.cacheKey = &cKey{
		src:        srcUsername,
		username:   scope.username,
		userId:     scope.userId,
		tenantName: scope.tenantName,
		tenantId:   scope.tenantId,
	}
}

func (b *managerBase) isSameUser(username, userId string, auth security.Authentication) bool {
	un, id, e := b.resolveUser(auth)
	if e != nil {
		return false
	}
	return username != "" && username == un || userId != "" && userId == id
}

func (b *managerBase) isSameTenant(tenantName, tenantId string, auth security.Authentication) bool {
	if tenantName == "" && tenantId == "" {
		return true
	}

	switch details := auth.Details().(type) {
	case security.TenantDetails:
		return tenantId != "" && tenantId == details.TenantId() || tenantName != "" && tenantName == details.TenantName()
	default:
		return false
	}
}

func (b *managerBase) cacheValidateFunc() validateFunc {
	return func(ctx context.Context, auth entryValue) bool {
		if auth == nil || auth.AccessToken() == nil {
			return false
		}

		_, e := b.tokenStoreReader.ReadAccessToken(ctx, auth.AccessToken().Value())
		return e == nil
	}
}

func (b *managerBase) cacheLoadFunc(rTime time.Time, authFunc authenticateFunc) loadFunc {
	return func(ctx context.Context, k cKey) (entryValue, time.Time, error) {
		auth, e := authFunc(ctx, &k)

		// calculate exp time based on backoff time
		errExp := rTime.UTC().Add(b.failureBackOff)
		if e != nil {
			return nil, b.calculateBackOffExp(e, errExp), e
		}

		if auth == nil {
			// sanity check, this shouldn't happen
			return nil, errExp, fmt.Errorf("[Internal Error] authenticateFunc returned nil oauth without error")
		}

		// try to guarantee token's validity by setting expire time a little earlier than auth's exp time
		oauth := auth.(oauth2.Authentication)
		tokenExp := oauth.AccessToken().ExpiryTime().UTC()
		exp := tokenExp.Add(-1 * b.guaranteedValidity)
		if exp.Before(rTime) {
			// edge case, we cannot guarantee token's validity, such error would insists until this token expires
			// we'd still return the token since it at least valid now,
			// but we set expire time to back-off time or token expiry, which ever is earlier
			if tokenExp.Before(errExp) {
				exp = tokenExp
			} else {
				exp = errExp
			}
		}

		return oauth, exp, nil
	}
}

func (b *managerBase) convertToAuthentication(ctx context.Context, result *seclient.Result) (oauth2.Authentication, error) {
	// TODO we could leverage IDToken and probably Remote token API
	auth, e := b.tokenStoreReader.ReadAuthentication(ctx, result.Token.Value(), oauth2.TokenHintAccessToken)
	if e != nil {
		return nil, e
	}

	// perform some checks
	switch {
	case auth.State() < security.StateAuthenticated:
		return nil, fmt.Errorf("token is not associated with an authenticated session")
	case auth.OAuth2Request().ClientId() == "":
		return nil, fmt.Errorf("token is not issued to a valid client")
	case auth.UserAuthentication() != nil && reflect.ValueOf(auth.UserAuthentication().Principal()).IsZero():
		return nil, fmt.Errorf("token is not authorized by a valid user")
	}
	return auth, nil
}

func (b *managerBase) calculateBackOffExp(err error, defaultValue time.Time) time.Time {
	switch {
	case errors.Is(err, httpclient.ErrorSubTypeDiscovery):
		return time.Now().UTC().Add(10 * time.Second)
	default:
		return defaultValue
	}
}