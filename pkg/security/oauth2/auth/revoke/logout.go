package revoke

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"net/http"
)

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

type HanlderOptions func(opt *HanlderOption)

type HanlderOption struct {
	Revoker auth.AccessRevoker
}

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
		h.handleDefault(ctx, auth)
	}
}

func (h TokenRevokingLogoutHanlder) handleGet(ctx context.Context, auth security.Authentication) {
	s := session.Get(ctx)
	if s == nil {
		return
	}

	_ = h.revoker.RevokeWithSessionId(ctx, s.GetID(), s.Name())
}

func (h TokenRevokingLogoutHanlder) handlePost(ctx context.Context, auth security.Authentication) {
	username, e := security.GetUsername(auth)
	if e != nil || username == "" {
		return
	}

	_ = h.revoker.RevokeWithUsername(ctx, username, true)
}

func (h TokenRevokingLogoutHanlder) handleDefault(ctx context.Context, auth security.Authentication) {
	// TODO implement me
}



