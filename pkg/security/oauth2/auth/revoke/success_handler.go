package revoke

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"net/http"
	"net/url"
)

type SuccessOptions func(opt *SuccessOption)

type SuccessOption struct {
	ClientStore         oauth2.OAuth2ClientStore
	RedirectWhitelist   utils.StringSet
	WhitelabelErrorPath string
}

// TokenRevokeSuccessHandler implements security.AuthenticationSuccessHandler
type TokenRevokeSuccessHandler struct {
	clientStore oauth2.OAuth2ClientStore
	whitelist   utils.StringSet
	fallback    security.AuthenticationErrorHandler
}

func NewTokenRevokeSuccessHandler(opts ...SuccessOptions) *TokenRevokeSuccessHandler {
	opt := SuccessOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &TokenRevokeSuccessHandler{
		clientStore: opt.ClientStore,
		fallback:    redirect.NewRedirectWithURL(opt.WhitelabelErrorPath),
		whitelist:   opt.RedirectWhitelist,
	}
}

func (h TokenRevokeSuccessHandler) HandleAuthenticationSuccess(ctx context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	switch r.Method {
	case http.MethodGet:
		fallthrough
	case http.MethodPost:
		h.redirect(ctx, r, rw, from, to)
	case http.MethodPut:
		fallthrough
	case http.MethodDelete:
		fallthrough
	default:
		h.status(ctx, rw)
	}
}

func (h TokenRevokeSuccessHandler) redirect(ctx context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	// Note: we don't have error handling alternatives (except for panic)
	redirectUri := r.FormValue(oauth2.ParameterRedirectUri)
	if redirectUri == "" {
		h.fallback.HandleAuthenticationError(ctx, r, rw, security.NewInternalError(fmt.Sprintf("missing %s", oauth2.ParameterRedirectUri)))
		return
	}

	clientId := r.FormValue(oauth2.ParameterClientId)
	client, e := auth.LoadAndValidateClientId(ctx, clientId, h.clientStore)
	if e != nil {
		h.fallback.HandleAuthenticationError(ctx, r, rw, e)
		return
	}

	resolved, e := auth.ResolveRedirectUri(ctx, redirectUri, client)
	if e != nil {
		// try resolve from whitelist
		if !h.isWhitelisted(ctx, redirectUri) {
			h.fallback.HandleAuthenticationError(ctx, r, rw, e)
			return
		}
		resolved = redirectUri
	}

	redirectUrl := h.appendWarnings(ctx, resolved)
	http.Redirect(rw, r, redirectUrl, http.StatusFound)
	_, _ = rw.Write([]byte{})
}

// In case of PUT, DELETE, PATCH etc, we don't clean authentication. Instead, we invalidate access token carried by header
func (h TokenRevokeSuccessHandler) status(_ context.Context, rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte{})
}

func (h TokenRevokeSuccessHandler) isWhitelisted(_ context.Context, redirect string) bool {
	for pattern, _ := range h.whitelist {
		matcher, e := auth.NewWildcardUrlMatcher(pattern)
		if e != nil {
			continue
		}
		if matches, e := matcher.Matches(redirect); e == nil && matches {
			return true
		}
	}
	return false
}

func (h TokenRevokeSuccessHandler) appendWarnings(ctx context.Context, redirect string) string {
	warnings := logout.GetWarnings(ctx)
	if len(warnings) == 0 {
		return redirect
	}

	redirectUrl, e := url.Parse(redirect)
	if e != nil {
		return redirect
	}

	q := redirectUrl.Query()
	for _, w := range warnings {
		q.Add("warning", w.Error())
	}
	redirectUrl.RawQuery = q.Encode()
	return redirectUrl.String()
}
