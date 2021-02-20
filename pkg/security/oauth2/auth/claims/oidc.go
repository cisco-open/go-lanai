package claims

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

func AuthenticationTime(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.AuthenticationDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.AuthenticationTime(), errorMissingDetails)
}

func FirstName(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.FirstName(), errorMissingDetails)
}

func LastName(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.LastName(), errorMissingDetails)
}

func Email(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.Email(), errorMissingDetails)
}

func Locale(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	details, ok := src.Details().(security.UserDetails)
	if !ok {
		return nil, errorMissingDetails
	}
	return nonZeroOrError(details.LocaleCode(), errorMissingDetails)
}
