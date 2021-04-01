package revoke

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"fmt"
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

func (h TokenRevokingLogoutHanlder) HandleLogout(ctx context.Context, r *http.Request, rw http.ResponseWriter, auth security.Authentication) error  {
	switch r.Method {
	case http.MethodGet:
		return h.handleGet(ctx, auth)
	case http.MethodPost:
		return h.handlePost(ctx, auth)
	case http.MethodPut:
		fallthrough
	case http.MethodDelete:
		return h.handleDefault(ctx, r)
	}
	return nil
}

func (h TokenRevokingLogoutHanlder) handleGet(ctx context.Context, auth security.Authentication) error {
	defer func() {
		security.Clear(ctx)
	}()
	s := session.Get(ctx)
	if s == nil {
		logger.WithContext(ctx).Debugf("invalid use of GET /logout endpoint. session is not found")
		return nil
	}

	if e := h.revoker.RevokeWithSessionId(ctx, s.GetID(), s.Name()); e != nil {
		logger.WithContext(ctx).Warnf("unable to revoke tokens with session %s: %v", s.GetID(), e)
		return e
	}
	return nil
}

func (h TokenRevokingLogoutHanlder) handlePost(ctx context.Context, auth security.Authentication) error  {
	defer func() {
		security.Clear(ctx)
	}()
	username, e := security.GetUsername(auth)
	if e != nil || username == "" {
		logger.WithContext(ctx).Debugf("invalid use of GET /logout endpoint. session is not found")
		return nil
	}

	if e := h.revoker.RevokeWithUsername(ctx, username, true); e != nil {
		logger.WithContext(ctx).Warnf("unable to revoke tokens with username %s: %v", username, e)
		return e
	}
	return nil
}

// In case of PUT, DELETE, PATCH etc, we don't clean authentication. Instead, we invalidate access token carried by header
func (h TokenRevokingLogoutHanlder) handleDefault(ctx context.Context, r *http.Request) error  {
	// grab token
	tokenValue, e := h.extractAccessToken(ctx, r)
	if e != nil {
		logger.WithContext(ctx).Warnf("unable to revoke token: %v", e)
		return nil
	}

	if e := h.revoker.RevokeWithTokenValue(ctx, tokenValue, auth.RevokerHintAccessToken); e != nil {
		logger.WithContext(ctx).Warnf("unable to revoke token with value %s: %v", log.Capped(tokenValue, 20), e)
		return e
	}
	return nil
}

func (h TokenRevokingLogoutHanlder) extractAccessToken(ctx context.Context, r *http.Request) (string, error) {
	// try header first
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, bearerTokenPrefix) {
		return strings.TrimPrefix(header, bearerTokenPrefix), nil
	}

	// then try param
	value := r.FormValue(oauth2.ParameterAccessToken)
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf(`access token is required either from "Authorization" header or parameter "%s"`, oauth2.ParameterAccessToken)
	}
	return value, nil
}


