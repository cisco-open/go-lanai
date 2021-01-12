package oauth2

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"errors"
)

// All "SubType" values are used as mask
// sub types of security.ErrorTypeCodeOAuth2
const (
	_                              = iota
	ErrorSubTypeCodeOAuth2Internal = security.ErrorTypeCodeOAuth2 + iota<<security.ErrorSubTypeOffset
	ErrorSubTypeCodeOAuth2Auth
	ErrorSubTypeCodeOAuth2Res
)

// ErrorSubTypeCodeOAuth2Internal
const (
	_                              = iota
	ErrorCodeOAuth2InternalGeneral = ErrorSubTypeCodeOAuth2Internal + iota
)

// ErrorSubTypeCodeOAuth2Auth
const (
	_ = iota
	ErrorCodeGranterNotAvalable = ErrorSubTypeCodeOAuth2Auth + iota
	ErrorCodeInvalidTokenRequest
)

// ErrorSubTypeCodeOAuth2Res
const (
	_ = iota
	ErrorCodeInvalidAccessToken = ErrorSubTypeCodeOAuth2Res + iota
)

// ErrorTypes, can be used in errors.Is
var (
	ErrorTypeOAuth2 = security.NewErrorType(security.ErrorTypeCodeOAuth2, errors.New("error type: oauth2"))

	ErrorSubTypeOAuth2Internal = security.NewErrorSubType(ErrorSubTypeCodeOAuth2Internal, errors.New("error sub-type: internal"))
	ErrorSubTypeOAuth2Auth     = security.NewErrorSubType(ErrorSubTypeCodeOAuth2Auth, errors.New("error sub-type: oauth2 auth"))
	ErrorSubTypeOAuth2Res      = security.NewErrorSubType(ErrorSubTypeCodeOAuth2Res, errors.New("error sub-type: oauth2 resource"))
)

/************************
	Constructors
*************************/
/* OAuth2Internal family */
func NewInternalError(text string, causes...interface{}) error {
	return security.NewCodedError(ErrorCodeOAuth2InternalGeneral, errors.New(text), causes...)
}

/* OAuth2Auth family */
func NewGranterNotAvailableError(value interface{}, causes...interface{}) error {
	return security.NewCodedError(ErrorCodeGranterNotAvalable, value, causes...)
}

func NewInvalidTokenRequestError(value interface{}, causes...interface{}) error {
	return security.NewCodedError(ErrorCodeInvalidTokenRequest, value, causes...)
}

/* OAuth2Res family */
func NewInvalidAccessTokenError(value interface{}, causes...interface{}) error {
	return security.NewCodedError(ErrorCodeInvalidAccessToken, value, causes...)
}