package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

var (
	errorInvalidRedirectUri = oauth2.NewInvalidRedirectUriError("")
)

// OAuth2ErrorHandler implements security.ErrorHandler
// It's responsible to handle all oauth2 errors
type OAuth2ErrorHandler struct {}

func NewOAuth2ErrorHanlder() *OAuth2ErrorHandler {
	return &OAuth2ErrorHandler{}
}

// security.ErrorHandler
func (h *OAuth2ErrorHandler) HandleError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	//err = h.translateError(c, err)
	h.handleError(c, r, rw, err)
}

func (h *OAuth2ErrorHandler) translateError(c context.Context, err error) error {
	switch {
	case errors.Is(err, oauth2.ErrorTypeOAuth2):
		return err
	case errors.Is(err, security.ErrorSubTypeUsernamePasswordAuth):
		return oauth2.NewInvalidClientError("invalid client", err)
	case errors.Is(err, security.ErrorTypeSecurity):
		return err
	default:
		return oauth2.NewInvalidGrantError(err.Error(), err)
	}
}

func (h *OAuth2ErrorHandler) handleError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {

	switch oe, ok := err.(oauth2.OAuth2ErrorTranslator); {
	case ok && errors.Is(err, oauth2.ErrorSubTypeOAuth2Internal):
		fallthrough
	case ok && errors.Is(err, oauth2.ErrorTypeOAuth2):
		// use redirect uri, fallback to standard error response
		tryWriteErrorAsRedirect(c, r, rw, oe)
	// No default, give other error handler chance to handle
	}
}

func writeOAuth2Error(c context.Context, r *http.Request, rw http.ResponseWriter, err oauth2.OAuth2ErrorTranslator) {
	challenge := ""
	sc := err.TranslateStatusCode()
	if sc == http.StatusUnauthorized || sc == http.StatusForbidden {
		challenge = fmt.Sprintf("%s %s", "Bearer", err.Error())
	}
	writeAdditionalHeader(c, r, rw, challenge)
	switch {
	case errors.Is(err, errorInvalidRedirectUri):
		security.WriteError(c, r, rw, sc, err)
	default:
		security.WriteErrorAsJson(c, rw, sc, err)
	}
}

func writeAdditionalHeader(c context.Context, r *http.Request, rw http.ResponseWriter, challenge string) {
	if security.IsResponseWritten(rw) {
		return
	}

	rw.Header().Add("Cache-Control", "no-store")
	rw.Header().Add("Pragma", "no-cache");

	if challenge != "" {
		rw.Header().Set("WWW-Authenticate", challenge);
	}
}

// given err have to be oauth2.OAuth2ErrorTranslator
func tryWriteErrorAsRedirect(c context.Context, r *http.Request, rw http.ResponseWriter, err oauth2.OAuth2ErrorTranslator) {
	if security.IsResponseWritten(rw) {
		return
	}

	urlStr, ok := findRedirectUri(c, r)
	if !ok {
		// fallback to default
		writeOAuth2Error(c, r, rw, err)
		return
	}

	params := map[string]string{}
	params[oauth2.ParameterError] = err.TranslateErrorCode()
	params[oauth2.ParameterErrorDescription] = err.Error()

	if state, ok := findRedirectState(c, r); ok {
		params[oauth2.ParameterState] = state
	}

	redirectUrl, e := appendRedirectUrl(urlStr, params)
	if e != nil {
		// fallback to default
		writeOAuth2Error(c, r, rw, err)
		return
	}
	http.Redirect(rw, r, redirectUrl, http.StatusFound)
}

func findRedirectUri(c context.Context, r *http.Request) (string, bool) {
	value, ok := c.Value(oauth2.CtxKeyResolvedAuthorizeRedirect).(string)
	return value, ok
}

func findRedirectState(c context.Context, r *http.Request) (string, bool) {
	if ar, e := findAuthorizeRequest(c, r); e == nil && ar.State != "" {
		return ar.RedirectUri, true
	}
	value, ok := c.Value(oauth2.CtxKeyResolvedAuthorizeState).(string)
	return value, ok
}

func findAuthorizeRequest(c context.Context, r *http.Request) (*AuthorizeRequest, error) {
	if ar, ok := c.Value(oauth2.CtxKeyValidatedAuthorizeRequest).(*AuthorizeRequest); ok {
		return ar, nil
	}

	if ar, ok := c.Value(oauth2.CtxKeyReceivedAuthorizeRequest).(*AuthorizeRequest); ok {
		return ar, nil
	}

	// last resort, try parse from request
	return ParseAuthorizeRequest(r)
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



