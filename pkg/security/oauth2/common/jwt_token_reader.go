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

func (r *jwtTokenStoreReader) ReadAuthentication(ctx context.Context, tokenValue string, hint oauth2.TokenHint) (oauth2.Authentication, error) {
	switch hint {
	case oauth2.TokenHintAccessToken:
		return r.readAuthenticationFromAccessToken(ctx, tokenValue)
	default:
		return nil, oauth2.NewUnsupportedTokenTypeError(fmt.Sprintf("token type [%s] is not supported", hint.String()))
	}
}

func (r *jwtTokenStoreReader) ReadAccessToken(c context.Context, value string) (oauth2.AccessToken, error) {
	token, e := r.parseAccessToken(c, value)
	switch {
	case e != nil:
		return nil, oauth2.NewInvalidAccessTokenError("token is invalid", e)
	case token.Expired():
		return nil, oauth2.NewInvalidAccessTokenError("token is expired")
	case !r.detailsStore.ContextDetailsExists(c, token):
		return nil, oauth2.NewInvalidAccessTokenError("token is revoked")
	}
	return token, nil
}

func (r *jwtTokenStoreReader) ReadRefreshToken(c context.Context, value string) (oauth2.RefreshToken, error) {
	token, e := r.parseRefreshToken(c, value)
	if e != nil {
		return nil, e
	}
	switch {
	case e != nil:
		return nil, oauth2.NewInvalidGrantError("refresh token is invalid", e)
	case token.WillExpire() && token.Expired():
		return nil, oauth2.NewInvalidGrantError("refresh token is expired")
	}
	return token, nil
}

func (r *jwtTokenStoreReader) parseAccessToken(c context.Context, value string) (*internal.DecodedAccessToken, error) {
	claims := internal.ExtendedClaims{}
	if e := r.jwtDecoder.DecodeWithClaims(c, value, &claims); e != nil {
		return nil, e
	}

	token := internal.DecodedAccessToken{}
	token.TokenValue = value
	token.DecodedClaims = &claims
	token.ExpireAt = claims.ExpiresAt
	token.IssuedAt = claims.IssuedAt
	token.ScopesSet = claims.Scopes.Copy()
	return &token, nil
}

func (r *jwtTokenStoreReader) parseRefreshToken(c context.Context, value string) (*internal.DecodedRefreshToken, error) {
	claims := internal.ExtendedClaims{}
	if e := r.jwtDecoder.DecodeWithClaims(c, value, &claims); e != nil {
		return nil, e
	}

	token := internal.DecodedRefreshToken{}
	token.TokenValue = value
	token.DecodedClaims = &claims
	token.ExpireAt = claims.ExpiresAt
	token.IssuedAt = claims.IssuedAt
	token.ScopesSet = claims.Scopes.Copy()
	return &token, nil
}

func (r *jwtTokenStoreReader) readAuthenticationFromAccessToken(c context.Context, tokenValue string) (oauth2.Authentication, error) {
	// parse JWT token
	token, e := r.parseAccessToken(c, tokenValue)
	if e != nil {
		return nil, e
	}

	claims := token.DecodedClaims
	if claims == nil {
		return nil, oauth2.NewInvalidAccessTokenError("token contains no claims")
	}

	// load context details
	details, e := r.detailsStore.ReadContextDetails(c, token)
	if e != nil {
		return nil, oauth2.NewInvalidAccessTokenError("token unknown", e)
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

/*****************
	Helpers
 *****************/
func (r *jwtTokenStoreReader) createOAuth2Request(claims *internal.ExtendedClaims) oauth2.OAuth2Request {
	clientId := claims.ClientId
	if clientId == "" && claims.Audience != nil && len(claims.Audience) != 0 {
		clientId = claims.Audience.Values()[0]
	}
	return oauth2.NewOAuth2Request(func(opt *oauth2.RequestDetails) {
		opt.Parameters = map[string]string{}
		opt.ClientId = clientId
		opt.Scopes = claims.Scopes
		opt.Approved = true
		opt.Extensions = claims.Values()
		//opt.GrantType =
		//opt.RedirectUri =
		//opt.ResponseTypes =
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
