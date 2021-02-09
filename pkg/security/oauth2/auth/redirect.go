package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
	"net/url"
)

/********************************************
	Helper functions for OAuth2 Redirects
 ********************************************/
func composeRedirectUrl(c context.Context, r *AuthorizeRequest, values map[string]string, useFragment bool) (string, error) {
	redirectUrl, ok := findRedirectUri(c, r)
	if !ok {
		return "", fmt.Errorf("redirect URI is unknown")
	}

	if state, ok := findRedirectState(c, r); ok {
		values[oauth2.ParameterState] = state
	}

	return appendRedirectUrl(redirectUrl, values)
}

func appendRedirectUrl(redirectUrl string, params map[string]string) (string, error) {
	loc, e := url.ParseRequestURI(redirectUrl)
	if e != nil || !loc.IsAbs() {
		return "", oauth2.NewInvalidRedirectUriError("invalid redirect_uri")
	}

	// TODO support fragments
	query := loc.Query()
	for k, v := range params {
		query.Add(k, v)
	}
	loc.RawQuery = query.Encode()

	loc.Redacted()
	return loc.String(), nil
}

func findRedirectUri(c context.Context, r *AuthorizeRequest) (string, bool) {
	value, ok := c.Value(oauth2.CtxKeyResolvedAuthorizeRedirect).(string)
	if !ok && r != nil {
		value = r.RedirectUri
		ok = true
	}
	return value, ok
}

func findRedirectState(c context.Context, r *AuthorizeRequest) (string, bool) {
	value, ok := c.Value(oauth2.CtxKeyResolvedAuthorizeState).(string)
	if !ok && r != nil {
		value = r.State
		ok = true
	}
	return value, ok
}



