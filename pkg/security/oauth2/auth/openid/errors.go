package openid

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"net/http"
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


