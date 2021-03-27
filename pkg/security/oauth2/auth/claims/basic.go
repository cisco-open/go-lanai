package claims

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"reflect"
	"time"
)

func ClientId(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.OAuth2Request() == nil {
		return nil, errorMissingRequest
	}
	return nonZeroOrError(opt.Source.OAuth2Request().ClientId(), errorMissingDetails)
}

func Audience(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.OAuth2Request() == nil {
		return nil, errorMissingRequest
	}
	if opt.Source.OAuth2Request().ClientId() == "" {
		return nil, errorMissingDetails
	}
	return utils.NewStringSet(opt.Source.OAuth2Request().ClientId()), nil
}

func JwtId(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	container, ok := opt.Source.AccessToken().(oauth2.ClaimsContainer)
	if !ok || container.Claims() == nil {
		return nil, errorMissingToken
	}

	return getClaim(container.Claims(), oauth2.ClaimJwtId)
}

func ExpiresAt(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.AccessToken() != nil {
		v = opt.Source.AccessToken().ExpiryTime()
	}

	if details, ok := opt.Source.Details().(security.ContextDetails); ok {
		v = details.ExpiryTime()
	}
	return nonZeroOrError(v, errorMissingDetails)
}

func IssuedAt(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.AccessToken() != nil {
		v = opt.Source.AccessToken().IssueTime()
	}

	if details, ok := opt.Source.Details().(security.ContextDetails); ok {
		v = details.IssueTime()
	}
	return nonZeroOrError(v, errorMissingDetails)
}

func Issuer(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Issuer != nil {
		if id := opt.Issuer.Identifier(); id != "" {
			return id, nil
		}
	}

	// fall back to extract from access token
	container, ok := opt.Source.AccessToken().(oauth2.ClaimsContainer)
	if !ok || container.Claims() == nil {
		return nil, errorMissingToken
	}

	return getClaim(container.Claims(), oauth2.ClaimIssuer)
}

func NotBefore(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	container, ok := opt.Source.AccessToken().(oauth2.ClaimsContainer)
	if !ok || container.Claims() == nil {
		return nil, errorMissingToken
	}

	return getClaim(container.Claims(), oauth2.ClaimNotBefore)
}

func Subject(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	return Username(ctx, opt)
}

func Scopes(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.OAuth2Request() == nil {
		return nil, errorMissingRequest
	}
	return nonZeroOrError(opt.Source.OAuth2Request().Scopes(), errorMissingDetails)
}

func Username(ctx context.Context, opt *FactoryOption) (v interface{}, err error) {
	if opt.Source.UserAuthentication() == nil || opt.Source.UserAuthentication().Principal() == nil {
		return nil, errorMissingUser
	}
	username, e := security.GetUsername(opt.Source.UserAuthentication())
	if e != nil {
		return nil, errorMissingUser
	}
	return nonZeroOrError(username, errorMissingDetails)
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