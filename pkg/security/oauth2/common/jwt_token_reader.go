package common

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common/internal"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"fmt"
)

// jwtTokenStoreReader implements TokenStoreReader
type jwtTokenStoreReader struct {
	detailsStore security.ContextDetailsStore
	jwtDecoder jwt.JwtDecoder
}

type JTSROptions func(opt *JTSROption)

type JTSROption struct {
	DetailsStore security.ContextDetailsStore
	Decoder jwt.JwtDecoder
}

func NewJwtTokenStoreReader(opts...JTSROptions) *jwtTokenStoreReader {
	opt := JTSROption{}
	for _, optFunc := range opts {
		optFunc(&opt)
	}

	return &jwtTokenStoreReader{
		detailsStore: opt.DetailsStore,
		jwtDecoder: opt.Decoder,
	}
}

func (r *jwtTokenStoreReader) ReadAuthentication(c context.Context, token oauth2.Token) (oauth2.Authentication, error) {
	switch token.(type) {
	case oauth2.AccessToken:
		return r.readAuthenticationFromAccessToken(c, token.(oauth2.AccessToken))
	case oauth2.RefreshToken:
		return r.readAuthenticationFromRefreshToken(c, token.(oauth2.AccessToken))
	default:
		return nil, oauth2.NewInternalError(fmt.Sprintf("token impl [%T] is not supported", token))
	}

}

func (r *jwtTokenStoreReader) ReadAccessToken(c context.Context, value string) (oauth2.AccessToken, error) {
	claims := internal.ExtendedClaims{}
	if e := r.jwtDecoder.DecodeWithClaims(c, value, &claims); e != nil {
		return nil, e
	}

	token := internal.DecodedAccessToken{}
	token.TokenValue = value
	token.Claims = &claims
	token.ExpireAt = claims.ExpiresAt
	token.IssuedAt = claims.IssuedAt
	token.ScopesSet = claims.Scopes.Copy()
	return &token, nil
}

func (r *jwtTokenStoreReader) ReadRefreshToken(c context.Context, value string) (oauth2.RefreshToken, error) {
	token := oauth2.NewDefaultRefreshToken(value)
	// TODO decode JWT
	return token, nil
}

func (r *jwtTokenStoreReader) readAuthenticationFromAccessToken(c context.Context, token oauth2.AccessToken) (oauth2.Authentication, error) {

	var claims *internal.ExtendedClaims
	switch token.(type) {
	case *internal.DecodedAccessToken:
		claims = token.(*internal.DecodedAccessToken).Claims
	case *oauth2.DefaultAccessToken:
		claims = internal.NewExtendedClaims(token.(*oauth2.DefaultAccessToken).Claims)
	default:
		return nil, oauth2.NewInternalError(fmt.Sprintf("token impl [%T] is not supported", token))
	}

	if claims == nil {
		return nil, oauth2.NewInvalidAccessTokenError("token contains no claims")
	}

	// load context details
	details, e := r.detailsStore.ReadContextDetails(c, token)
	if e != nil {
		return nil, e
	}

	// reconstruct request
	request := r.createOAuth2Request(claims)

	// reconstruct user auth if available
	var userAuth security.Authentication
	if claims.Subject != "" {
		userAuth = r.createUserAuthentication(claims, details)
	}

	return oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
		opt.Request = request
		opt.UserAuth = userAuth
		opt.Token = token
		opt.Details = details
	}), nil
}

func (r *jwtTokenStoreReader) readAuthenticationFromRefreshToken(c context.Context, token oauth2.AccessToken) (oauth2.Authentication, error) {
	// TODO implement me
	panic("implement me")
}

/*****************
	Helpers
 *****************/
func (r *jwtTokenStoreReader) createOAuth2Request(claims *internal.ExtendedClaims) oauth2.OAuth2Request {
	return oauth2.NewOAuth2Request(func(opt *oauth2.RequestDetails) {
		opt.Parameters = map[string]string{}
		opt.ClientId = claims.ClientId
		opt.Scopes = claims.Scopes
		opt.Approved = true
		//opt.GrantType =
		//opt.RedirectUri =
		//opt.ResponseTypes =
		//opt.Extensions =
	})
}

func (r *jwtTokenStoreReader) createUserAuthentication(claims *internal.ExtendedClaims, details security.ContextDetails) security.Authentication {
	permissions := map[string]interface{}{}
	for k, _ := range details.Permissions() {
		permissions[k] = true
	}

	return oauth2.NewUserAuthentication(func(opt *oauth2.UserAuthOption) {
		opt.Principal = claims.Subject
		opt.Permissions = permissions
		opt.State = security.StateAuthenticated
		// TODO maybe support extra fields from claims
		opt.Details = map[string]interface{}{}
	})
}
