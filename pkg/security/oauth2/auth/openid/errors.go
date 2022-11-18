package openid

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	errorutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/error"
	"errors"
	"net/http"
)

const (
	_ = iota
	// ErrorSubTypeCodeOidcSlo non-programming error that can occur during oidc RP initiated logout
	ErrorSubTypeCodeOidcSlo = security.ErrorTypeCodeOidc + iota<<errorutils.ErrorSubTypeOffset
)

const (
	_ = ErrorSubTypeCodeOidcSlo + iota
	ErrorCodeOidcSloRp
	ErrorCodeOidcSloOp
)

var (
	ErrorSubTypeOidcSlo = security.NewErrorSubType(ErrorSubTypeCodeOidcSlo, errors.New("error sub-type: oidc slo"))

	// ErrorOidcSloRp errors are displayed as an HTML page with status 400
	ErrorOidcSloRp = security.NewCodedError(ErrorCodeOidcSloRp, "SLO rp error")
	// ErrorOidcSloOp errors are displayed as an HTML page with status 500
	ErrorOidcSloOp = security.NewCodedError(ErrorCodeOidcSloOp, "SLO op error")
)

func newOpenIDExtendedError(oauth2Code string, value interface{}, causes []interface{}) error {
	return oauth2.NewOAuth2Error(oauth2.ErrorCodeOpenIDExt, value,
		oauth2Code, http.StatusBadRequest, causes...)
}

func NewOpenIDExtendedError(oauth2Code string, value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2Code, value, causes)
}

func NewInteractionRequiredError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationInteractionRequired, value, causes)
}

func NewLoginRequiredError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationLoginRequired, value, causes)
}

func NewAccountSelectionRequiredError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationAcctSelectRequired, value, causes)
}

func NewInvalidRequestURIError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationInvalidRequestURI, value, causes)
}

func NewInvalidRequestObjError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationInvalidRequestObj, value, causes)
}

func NewRequestNotSupportedError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationRequestUnsupported, value, causes)
}

func NewRequestURINotSupportedError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationRequestURIUnsupported, value, causes)
}

func NewRegistrationNotSupportedError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationRegistrationUnsupported, value, causes)
}
