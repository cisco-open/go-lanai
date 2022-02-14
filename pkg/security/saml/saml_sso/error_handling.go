package saml_auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"errors"
	"github.com/crewjam/saml"
	"net/http"
)

const CtxKeySamlAuthnRequest = "kSamlAuthnRequest"

type SamlErrorHandler struct {

}

func NewSamlErrorHandler() *SamlErrorHandler {
	return &SamlErrorHandler{}
}

// HandleError
/**
Handles error as saml response when possible.
Otherwise let the error handling handle it

See http://docs.oasis-open.org/security/saml/v2.0/saml-profiles-2.0-os.pdf 4.1.3.5
*/
//nolint:errorlint
func (h *SamlErrorHandler) HandleError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	//catch the saml errors that were weren't able to send back to the client
	if !errors.Is(err, security.ErrorTypeSaml) {
		return
	}

	authRequest, ok := c.Value(CtxKeySamlAuthnRequest).(*saml.IdpAuthnRequest)

	if errors.Is(err, ErrorSubTypeSamlInternal) || !ok {
		writeErrorAsHtml(c, r, rw, err)
	} else if errors.Is(err, ErrorSubTypeSamlSso) {
		code := saml.StatusResponder
		message := ""
		if translator, ok := err.(SamlSsoErrorTranslator); ok { //all the saml sub types should implement the translator API
			code = translator.TranslateErrorCode()
			message = translator.TranslateErrorMessage()
		}
		respErr := MakeErrorResponse(authRequest, code, message)
		if respErr != nil {
			writeErrorAsHtml(c, r, rw, NewSamlInternalError("cannot create response", respErr))
		}
		writeErr := authRequest.WriteResponse(rw)
		if writeErr != nil {
			writeErrorAsHtml(c, r, rw, NewSamlInternalError("cannot write response", writeErr))
		}
	}
}

func writeErrorAsHtml(c context.Context, _ *http.Request, rw http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	//nolint:errorlint
	if translator, ok := err.(SamlSsoErrorTranslator); ok { //all the saml errors should implement this interface
		code = translator.TranslateHttpStatusCode()
	}
	security.WriteErrorAsHtml(c, rw, code, err)
}