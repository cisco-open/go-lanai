package revoke

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
)

type RevokerOptions func(opt *RevokerOption)

type RevokerOption struct {
	AuthRegistry auth.AuthorizationRegistry
	SessionStore session.Store
}

// DefaultAccessRevoker impelments auth.AccessRevoker
type DefaultAccessRevoker struct {
	authRegistry auth.AuthorizationRegistry
	sessionStore session.Store
}

func NewDefaultAccessRevoker(opts...RevokerOptions) *DefaultAccessRevoker {
	opt := RevokerOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &DefaultAccessRevoker{
		authRegistry: opt.AuthRegistry,
		sessionStore: opt.SessionStore,
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
		logger.WithContext(ctx).Warnf("Unable to revoke tokens for session ID [%s]: %v", sessionId, e)
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
		logger.WithContext(ctx).Warnf("Unable to revoke tokens for username [%s]: %v", username, e)
		return e
	}
	return
}

func (r DefaultAccessRevoker) RevokeWithClientId(ctx context.Context, clientId string, revokeRefreshToken bool) error {
	// revoke all tokens
	if e := r.authRegistry.RevokeClientAccess(ctx, clientId, true); e != nil {
		logger.WithContext(ctx).Warnf("Unable to revoke tokens for client [%s]: %v", clientId, e)
		return e
	}
	return nil
}

func (r DefaultAccessRevoker) RevokeRefreshToken(ctx context.Context, token oauth2.RefreshToken) error {
	// TODO implement me
	panic("implement me")
}

func (r DefaultAccessRevoker) RevokeAccessToken(ctx context.Context, token oauth2.AccessToken) error {
	// TODO implement me
	panic("implement me")
}

func (r DefaultAccessRevoker) RevokeAllAccessTokens(ctx context.Context, token oauth2.RefreshToken) error {
	// TODO implement me
	panic("implement me")
}

