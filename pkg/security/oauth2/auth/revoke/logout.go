package revoke

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"net/http"
	"strings"
)

const (
	bearerTokenPrefix = "Bearer "
)

type HanlderOptions func(opt *HanlderOption)

type HanlderOption struct {
	Revoker auth.AccessRevoker
}

/**
 * TokenRevokingLogoutHanlder
 *
 * GET method: used for logout by the session controlled clients. The client send user to this endpoint and the session
 * is invalidated. As a result, the tokens controlled by this session is invalidated (See the NfvClientDetails.useSessionTimeout
 * properties). In addition, if an access token is passed in the request, the access token will be invalidated explicitly.
 *
 * POST method: used for SSO logout. Typically browser based. The client redirect user to this endpoint and
 * we revoke all tokens associated with this user
 *
 * PUT/DELETE method: used for token revocation. Typically for service login or token revocation. We grab token
 * from header and revoke this only this token.
 *
 * @author Livan Du
 * Created on 2018-05-04
 */
type TokenRevokingLogoutHanlder struct {
	revoker auth.AccessRevoker
}

func NewTokenRevokingLogoutHanlder(opts...HanlderOptions) *TokenRevokingLogoutHanlder {
	opt := HanlderOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &TokenRevokingLogoutHanlder{
		revoker: opt.Revoker,
	}
}

func (h TokenRevokingLogoutHanlder) HandleLogout(ctx context.Context, r *http.Request, rw http.ResponseWriter, auth security.Authentication) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(ctx, auth)
	case http.MethodPost:
		h.handlePost(ctx, auth)
	case http.MethodPut:
		fallthrough
	case http.MethodDelete:
		fallthrough
	default:
		h.handleDefault(ctx, r)
	}
}

func (h TokenRevokingLogoutHanlder) handleGet(ctx context.Context, auth security.Authentication) {
	s := session.Get(ctx)
	if s == nil {
		logger.WithContext(ctx).Debugf("invalid use of GET /logout endpoint. session is not found")
		return
	}

	if e := h.revoker.RevokeWithSessionId(ctx, s.GetID(), s.Name()); e != nil {
		logger.WithContext(ctx).Warnf("unable to revoke tokens with session %s: %v", s.GetID(), e)
	}
	security.Clear(ctx)
}

func (h TokenRevokingLogoutHanlder) handlePost(ctx context.Context, auth security.Authentication) {
	username, e := security.GetUsername(auth)
	if e != nil || username == "" {
		logger.WithContext(ctx).Debugf("invalid use of GET /logout endpoint. session is not found")
		return
	}

	if e := h.revoker.RevokeWithUsername(ctx, username, true); e != nil {
		logger.WithContext(ctx).Warnf("unable to revoke tokens with username %s: %v", username, e)
	}
	security.Clear(ctx)
}

// In case of PUT, DELETE, PATCH etc, we don't clean authentication. Instead, we invalidate access token carried by header
func (h TokenRevokingLogoutHanlder) handleDefault(ctx context.Context, r *http.Request) {
	// grab bearer token // TODO also extract token from parameter
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, bearerTokenPrefix) {
		return
	}

	tokenValue := strings.TrimPrefix(header, bearerTokenPrefix)
	if e := h.revoker.RevokeWithTokenValue(ctx, tokenValue, auth.RevokerHintAccessToken); e != nil {
		logger.WithContext(ctx).Warnf("unable to revoke token with value %s: %v", log.Capped(tokenValue, 20), e)
	}
}



