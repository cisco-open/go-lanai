package revoke

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"fmt"
)

type RevokerOptions func(opt *RevokerOption)

type RevokerOption struct {
	AuthRegistry     auth.AuthorizationRegistry
	SessionStore     session.Store
	TokenStoreReader oauth2.TokenStoreReader
}

// DefaultAccessRevoker impelments auth.AccessRevoker
type DefaultAccessRevoker struct {
	authRegistry     auth.AuthorizationRegistry
	sessionStore     session.Store
	tokenStoreReader oauth2.TokenStoreReader
}

func NewDefaultAccessRevoker(opts ...RevokerOptions) *DefaultAccessRevoker {
	opt := RevokerOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &DefaultAccessRevoker{
		authRegistry:     opt.AuthRegistry,
		sessionStore:     opt.SessionStore,
		tokenStoreReader: opt.TokenStoreReader,
	}
}

func (r DefaultAccessRevoker) RevokeWithSessionId(ctx context.Context, sessionId string, sessionName string) error {
	// expire session
	if s, e := r.sessionStore.Get(sessionId, sessionName); e == nil && s != nil {
		if e := r.sessionStore.WithContext(ctx).Delete(s); e != nil {
			logger.WithContext(ctx).Warnf("Unable to expire session for session ID [%s]: %v", s.GetID(), e)
		}
	}

	// revoke all tokens
	if e := r.authRegistry.RevokeSessionAccess(ctx, sessionId, true); e != nil {
		return e
	}
	return nil
}

func (r DefaultAccessRevoker) RevokeWithUsername(ctx context.Context, username string, revokeRefreshToken bool) (err error) {
	// expire all sessions
	if sessions, e := r.sessionStore.FindByPrincipalName(username, session.DefaultName); e == nil {
		for _, s := range sessions {
			if e := r.sessionStore.WithContext(ctx).Delete(s); e != nil {
				logger.WithContext(ctx).Warnf("Unable to expire session for session ID [%s]: %v", s.GetID(), e)
				err = e
			}
		}
	} else {
		err = e
	}

	// revoke all tokens
	if e := r.authRegistry.RevokeUserAccess(ctx, username, true); e != nil {
		return e
	}
	return
}

func (r DefaultAccessRevoker) RevokeWithClientId(ctx context.Context, clientId string, revokeRefreshToken bool) error {
	return r.authRegistry.RevokeClientAccess(ctx, clientId, true)
}

func (r DefaultAccessRevoker) RevokeWithTokenValue(ctx context.Context, tokenValue string, hint auth.RevokerTokenHint) error {
	switch hint {
	case auth.RevokerHintAccessToken:
		token, e := r.tokenStoreReader.ReadAccessToken(ctx, tokenValue)
		if e != nil {
			return e
		}
		return r.authRegistry.RevokeAccessToken(ctx, token)
	case auth.RevokerHintRefreshToken:
		token, e := r.tokenStoreReader.ReadRefreshToken(ctx, tokenValue)
		if e != nil {
			return e
		}
		return r.authRegistry.RevokeRefreshToken(ctx, token)
	default:
		return fmt.Errorf("unsupported revoker token hint")
	}
}
