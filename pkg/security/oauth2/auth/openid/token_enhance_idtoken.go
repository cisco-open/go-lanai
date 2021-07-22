package openid

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
)

/*****************************
	ID Token Enhancer
 *****************************/

var (
	scopedSpecs = map[string]map[string]claims.ClaimSpec{
		oauth2.ScopeOidcProfile: claims.ProfileScopeSpecs,
		oauth2.ScopeOidcEmail: claims.EmailScopeSpecs,
		oauth2.ScopeOidcPhone: claims.PhoneScopeSpecs,
		oauth2.ScopeOidcAddress: claims.AddressScopeSpecs,
	}
	defaultSpecs = []map[string]claims.ClaimSpec{
		claims.IdTokenBasicSpecs,
	}
	fullSpecs = make([]map[string]claims.ClaimSpec, len(defaultSpecs), len(scopedSpecs) + len(defaultSpecs))

)

func init() {
	for i, specs := range defaultSpecs {
		fullSpecs[i] = specs
	}
	for _, specs := range scopedSpecs {
		fullSpecs = append(fullSpecs, specs)
	}
}

type EnhancerOptions func(opt *EnhancerOption)
type EnhancerOption struct {
	Issuer       security.Issuer
	JwtEncoder   jwt.JwtEncoder
}

// OpenIDTokenEnhancer implements order.Ordered and TokenEnhancer
// OpenIDTokenEnhancer generate OpenID ID Token and set it to token details
//goland:noinspection GoNameStartsWithPackageName
type OpenIDTokenEnhancer struct {
	issuer       security.Issuer
	jwtEncoder   jwt.JwtEncoder
}

func NewOpenIDTokenEnhancer(opts ...EnhancerOptions) *OpenIDTokenEnhancer {
	opt := EnhancerOption{}
	for _, fn := range opts {
		fn(&opt)
	}
	return &OpenIDTokenEnhancer{
		issuer:     opt.Issuer,
		jwtEncoder: opt.JwtEncoder,
	}
}

func (oe *OpenIDTokenEnhancer) Order() int {
	return auth.TokenEnhancerOrderTokenDetails
}

func (oe *OpenIDTokenEnhancer) Enhance(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	if oe.shouldSkip(oauth) {
		return token, nil
	}

	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, oauth2.NewInternalError("unsupported token implementation %T", t)
	}

	specs := oe.determineClaimSpecs(oauth.OAuth2Request())
	c := IdTokenClaims{}
	e := claims.Populate(ctx, &c,
		claims.WithSpecs(specs...),
		claims.WithSource(oauth),
		claims.WithIssuer(oe.issuer),
		claims.WithAccessToken(token),
	)

	if e != nil {
		return nil, oauth2.NewInternalError(e)
	}

	idToken, e := oe.jwtEncoder.Encode(ctx, &c)
	if e != nil {
		return nil, oauth2.NewInternalError(e)
	}

	t.PutDetails(oauth2.JsonFieldIDTokenValue, idToken)

	return t, nil
}

func (oe *OpenIDTokenEnhancer) shouldSkip(oauth oauth2.Authentication) bool {
	req := oauth.OAuth2Request()
	return req == nil ||
		// grant type not supported
		!SupportedGrantTypes.Has(req.GrantType()) ||
		// openid scope not requested
		!req.Scopes().Has(oauth2.ScopeOidc) ||
		// not user authorized
		oauth.UserAuthentication() == nil
}

func (oe *OpenIDTokenEnhancer) determineClaimSpecs(request oauth2.OAuth2Request) []map[string]claims.ClaimSpec {
	if request == nil || request.Scopes() == nil || !request.Approved() {
		return defaultSpecs
	}

	specs := make([]map[string]claims.ClaimSpec, len(defaultSpecs), len(defaultSpecs) + len(request.Scopes()))
	for i, spec := range defaultSpecs {
		specs[i] = spec
	}

	scopes := request.Scopes()
	includeAll := true
	for scope, spec := range scopedSpecs {
		if scopes.Has(scope) {
			specs = append(specs, spec)
			includeAll = false
		}
	}

	if includeAll {
		return fullSpecs
	}
	return specs
}

