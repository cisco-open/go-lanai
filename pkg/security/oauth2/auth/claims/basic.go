package claims

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"reflect"
	"time"
)

func ClientId(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	if src.OAuth2Request() == nil {
		return nil, errorMissingRequest
	}
	return nonZeroOrError(src.OAuth2Request().ClientId(), errorMissingDetails)
}

func Audience(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	if src.OAuth2Request() == nil {
		return nil, errorMissingRequest
	}
	if src.OAuth2Request().ClientId() == "" {
		return nil, errorMissingDetails
	}
	return utils.NewStringSet(src.OAuth2Request().ClientId()), nil
}

func JwtId(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	container, ok := src.AccessToken().(oauth2.ClaimsContainer)
	if !ok || container.Claims() == nil {
		return nil, errorMissingToken
	}

	return getClaim(container.Claims(), oauth2.ClaimJwtId)
}

func ExpiresAt(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	if src.AccessToken() != nil {
		v = src.AccessToken().ExpiryTime()
	}

	if details, ok := src.Details().(security.ContextDetails); ok {
		v = details.ExpiryTime()
	}
	return nonZeroOrError(v, errorMissingDetails)
}

func IssuedAt(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	if src.AccessToken() != nil {
		v = src.AccessToken().IssueTime()
	}

	if details, ok := src.Details().(security.ContextDetails); ok {
		v = details.IssueTime()
	}
	return nonZeroOrError(v, errorMissingDetails)
}

func Issuer(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	container, ok := src.AccessToken().(oauth2.ClaimsContainer)
	if !ok || container.Claims() == nil {
		return nil, errorMissingToken
	}

	return getClaim(container.Claims(), oauth2.ClaimIssuer)
}

func NotBefore(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	container, ok := src.AccessToken().(oauth2.ClaimsContainer)
	if !ok || container.Claims() == nil {
		return nil, errorMissingToken
	}

	return getClaim(container.Claims(), oauth2.ClaimNotBefore)
}

func Subject(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	return Username(ctx, src)
}

func Scopes(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	if src.OAuth2Request() == nil {
		return nil, errorMissingRequest
	}
	return nonZeroOrError(src.OAuth2Request().Scopes(), errorMissingDetails)
}

func Username(ctx context.Context, src oauth2.Authentication) (v interface{}, err error) {
	if src.UserAuthentication() == nil || src.UserAuthentication().Principal() == nil {
		return nil, errorMissingUser
	}
	principal := src.UserAuthentication().Principal()
	switch principal.(type) {
	case string:
		v = principal.(string)
	case fmt.Stringer:
		v = principal.(fmt.Stringer).String()
	case security.Account:
		v = principal.(security.Account).Username()
	default:
		return nil, errorMissingUser
	}
	return nonZeroOrError(v, errorMissingDetails)
}

func nonZeroOrError(v interface{}, candidateError error) (interface{}, error) {
	var isZero bool
	switch v.(type) {
	case string:
		isZero = v.(string) == ""
	case time.Time:
		isZero = v.(time.Time).IsZero()
	case utils.StringSet:
		isZero = len(v.(utils.StringSet)) == 0
	default:
		isZero = reflect.ValueOf(v).IsZero()
	}

	if isZero {
		return nil, candidateError
	}
	return v, nil
}

func getClaim(claims oauth2.Claims, claim string) (v interface{}, err error) {
	if !claims.Has(claim) {
		return nil, errorMissingClaims
	}
	return claims.Get(claim), nil
}