package saml_auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"errors"
	"fmt"
	"github.com/crewjam/saml"
	"net/http"
)

const CtxKeySamlAuthnRequest = "kSamlAuthnRequest"

type SamlErrorHandler struct {}

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

	switch {
	case errors.Is(err, ErrorSubTypeSamlInternal):
		writeErrorAsHtml(c, r, rw, err)
	case errors.Is(err, ErrorSubTypeSamlSso):
		h.handleSsoError(c, r, rw, err)
	case errors.Is(err, ErrorSubTypeSamlSlo):
		h.handleSloError(c, r, rw, err)
	}
}

//nolint:errorlint
func (h *SamlErrorHandler) handleSsoError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	authRequest, ok := c.Value(CtxKeySamlAuthnRequest).(*saml.IdpAuthnRequest)
	if !ok {
		writeErrorAsHtml(c, r, rw, err)
	}

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

//nolint:errorlint
func (h *SamlErrorHandler) handleSloError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	sloRequest, ok := c.Value(ctxKeySloRequest).(*SamlLogoutRequest)
	if !ok {
		writeErrorAsHtml(c, r, rw, err)
		return
	}

	code := saml.StatusAuthnFailed
	message := err.Error()
	if translator, ok := err.(SamlSsoErrorTranslator); ok { //all the saml sub types should implement the translator API
		code = translator.TranslateErrorCode()
		message = translator.TranslateErrorMessage()
	}

	switch {
	case errors.Is(err, ErrorSamlSloRequester):
		// requester error, means requester is not validated, we display errors as HTML
		writeErrorAsHtml(c, r, rw, err)
		return
	}

	resp, e := MakeLogoutResponse(sloRequest, code, message)
	if e != nil {
		msg := fmt.Sprintf("unable to create logout error response with code [%s]: %s. Reason: %v", code, message, e)
		writeErrorAsHtml(c, r, rw, NewSamlInternalError(msg, e))
		return
	}
	sloRequest.Response = resp
	if e := sloRequest.WriteResponse(rw); e != nil {
		msg := fmt.Sprintf("unable to send logout error response with code [%s]: %s. Reason: %v", code, message, e)
		writeErrorAsHtml(c, r, rw, NewSamlInternalError(msg, e))
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