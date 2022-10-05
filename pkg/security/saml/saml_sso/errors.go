package saml_auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	errorutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/error"
	"errors"
	"github.com/crewjam/saml"
	"net/http"
)

//errors maps to the status code described in section 3.2.2 of http://docs.oasis-open.org/security/saml/v2.0/saml-core-2.0-os.pdf

const (
	_ = iota
	// ErrorSubTypeCodeSamlSso non-programming error that can occur during SAML web sso flow. These errors will be returned to the requester
	// as a status code when possible
	ErrorSubTypeCodeSamlSso = security.ErrorTypeCodeSaml + iota<<errorutils.ErrorSubTypeOffset
	// ErrorSubTypeCodeSamlSlo non-programming error that can occur during SAML SLO flow
	ErrorSubTypeCodeSamlSlo
	// ErrorSubTypeCodeSamlInternal programming error, these will be displayed on an error page
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

// ErrorSubTypeCodeSamlSlo
const (
	_ = ErrorSubTypeCodeSamlSlo + iota
	ErrorCodeSamlSloRequester
	ErrorCodeSamlSloResponder
)

// ErrorSubTypeCodeSamlInternal
const (
	_ = ErrorSubTypeCodeSamlInternal + iota
	ErrorCodeSamlInternalGeneral
)

var (
	ErrorSubTypeSamlSso      = security.NewErrorSubType(ErrorSubTypeCodeSamlSso, errors.New("error sub-type: sso"))
	ErrorSubTypeSamlSlo      = security.NewErrorSubType(ErrorSubTypeCodeSamlSlo, errors.New("error sub-type: slo"))
	ErrorSubTypeSamlInternal = security.NewErrorSubType(ErrorSubTypeCodeSamlInternal, errors.New("error sub-type: internal"))

	// ErrorSamlSloRequester requester errors are displayed as a HTML page
	ErrorSamlSloRequester = security.NewCodedError(ErrorCodeSamlSloRequester, "SLO requester error")
	// ErrorSamlSloResponder responder errors are communicated back to SP via bindings
	ErrorSamlSloResponder = security.NewCodedError(ErrorCodeSamlSloResponder, "SLO responder error")
)

type SamlSsoErrorTranslator interface {
	error
	TranslateErrorCode() string
	TranslateErrorMessage() string
	TranslateHttpStatusCode() int
}

type SamlError struct {
	security.CodedError
	EC string // saml error code
	SC int    // status code
}

func NewSamlError(code int64, e interface{}, samlErrorCode string, httpStatusCode int, causes ...interface{}) *SamlError {
	embedded := security.NewCodedError(code, e, causes...)
	return &SamlError{
		CodedError: *embedded,
		EC:         samlErrorCode,
		SC:         httpStatusCode,
	}
}

func (s *SamlError) TranslateErrorCode() string {
	return s.EC
}

func (s *SamlError) TranslateErrorMessage() string {
	return s.Error()
}

func (s *SamlError) TranslateHttpStatusCode() int {
	return s.SC
}

func NewSamlInternalError(text string, causes ...interface{}) error {
	return NewSamlError(ErrorCodeSamlInternalGeneral, errors.New(text), "", http.StatusInternalServerError, causes...)
}

func NewSamlRequesterError(text string, causes ...interface{}) error {
	return NewSamlError(ErrorCodeSamlSsoRequester, errors.New(text), saml.StatusRequester, http.StatusBadRequest, causes...)
}

func NewSamlResponderError(text string, causes ...interface{}) error {
	return NewSamlError(ErrorCodeSamlSsoResponder, errors.New(text), saml.StatusResponder, http.StatusInternalServerError, causes...)
}

func NewSamlRequestVersionMismatch(text string, causes ...interface{}) error {
	return NewSamlError(ErrorCodeSamlSsoRequestVersionMismatch, errors.New(text), saml.StatusVersionMismatch, http.StatusConflict, causes...)
}
