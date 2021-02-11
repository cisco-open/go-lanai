package tokenauth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"errors"
	"fmt"
	"net/http"
)

// OAuth2ErrorHandler implements security.ErrorHandler
// It's responsible to handle all oauth2 errors
type OAuth2ErrorHandler struct {}

func NewOAuth2ErrorHanlder() *OAuth2ErrorHandler {
	return &OAuth2ErrorHandler{}
}

// security.ErrorHandler
func (h *OAuth2ErrorHandler) HandleError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	h.handleError(c, r, rw, err)
}

func (h *OAuth2ErrorHandler) handleError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {

	switch oe, ok := err.(oauth2.OAuth2ErrorTranslator); {
	case ok && errors.Is(err, oauth2.ErrorTypeOAuth2):
		writeOAuth2Error(c, r, rw, oe)
	}
}

func writeOAuth2Error(c context.Context, r *http.Request, rw http.ResponseWriter, err oauth2.OAuth2ErrorTranslator) {
	challenge := ""
	sc := err.TranslateStatusCode()
	if sc == http.StatusUnauthorized || sc == http.StatusForbidden {
		challenge = fmt.Sprintf("%s %s", "Bearer", err.Error())
	}
	writeAdditionalHeader(c, r, rw, challenge)
	security.WriteError(c, r, rw, sc, err)
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

