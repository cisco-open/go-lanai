package saml_auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"errors"
)

//errors maps to the status code described in section 3.2.2 of http://docs.oasis-open.org/security/saml/v2.0/saml-core-2.0-os.pdf

const (
	_                              = iota
	// non-programming error that can occur during SAML web sso flow. These errors will be returned to the requester
	// as a status code when possible
	ErrorSubTypeCodeSamlSso = security.ErrorTypeCodeSaml + iota<<security.ErrorSubTypeOffset
	//programming error, these will be displayed on an error page
	// so that we can fix the error on our end.
	ErrorSubTypeCodeSamlInternal
)

// ErrorSubTypeCodeSamlSso
const (
	_ = ErrorSubTypeCodeSamlSso + iota
	ErrorCodeSamlSsoRequester
	ErrorCodeSamlSsoResponder
	ErrorCodeSamlSsoRequestVersionMismatch
)

// ErrorSubTypeCodeSamlInternal
const (
	_ = ErrorSubTypeCodeSamlInternal + iota
	ErrorCodeSamlInternalGeneral
)

var (
	ErrorSubTypeSamlSso   = security.NewErrorSubType(ErrorSubTypeCodeSamlSso, errors.New("error sub-type: sso"))
	ErrorSubTypeSamlInternal = security.NewErrorSubType(ErrorSubTypeCodeSamlInternal, errors.New("error sub-type: internal"))
)

//TODO: if we need error handler to write to the service provider, we'll need some additional information here
type SamlError struct {
	security.CodedError
}

func NewSamlError(code int, e interface{}, causes...interface{}) *SamlError {
	embedded := security.NewCodedError(code, e, causes...)
	return &SamlError{
		CodedError:  *embedded,
	}
}

func NewSamlInternalError(text string, causes...interface{}) error {
	return NewSamlError(ErrorCodeSamlInternalGeneral, errors.New(text), causes...)
}

func NewSamlRequesterError(text string, causes...interface{}) error {
	return NewSamlError(ErrorCodeSamlSsoRequester, errors.New(text), causes...)
}

func NewSamlResponderError(text string, causes...interface{}) error {
	return NewSamlError(ErrorCodeSamlSsoResponder, errors.New(text), causes...)
}

func NewSamlRequestVersionMismatch(text string, causes...interface{}) error {
	return NewSamlError(ErrorCodeSamlSsoRequestVersionMismatch, errors.New(text), causes...)
}

type SamlSsoErrorTranslator interface {
	error
	TranslateErrorCode() string
	TranslateErrorMessage() string
	TranslateHttpStatusCode() int
}
