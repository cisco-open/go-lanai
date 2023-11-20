package misc

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/openid"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/json"
	httptransport "github.com/go-kit/kit/transport/http"
)

var (
	scopedSpecs = map[string]map[string]claims.ClaimSpec{
		oauth2.ScopeOidcProfile: claims.ProfileScopeSpecs,
		oauth2.ScopeOidcEmail:   claims.EmailScopeSpecs,
		oauth2.ScopeOidcPhone:   claims.PhoneScopeSpecs,
		oauth2.ScopeOidcAddress: claims.AddressScopeSpecs,
	}
	defaultSpecs = []map[string]claims.ClaimSpec{
		claims.UserInfoBasicSpecs,
	}
	fullSpecs = []map[string]claims.ClaimSpec{
		claims.UserInfoBasicSpecs,
		claims.ProfileScopeSpecs,
		claims.EmailScopeSpecs,
		claims.PhoneScopeSpecs,
		claims.AddressScopeSpecs,
	}
)

type UserInfoRequest struct{}

type UserInfoPlainResponse struct {
	UserInfoClaims
}

type UserInfoJwtResponse string

// MarshalText implements encoding.TextMarshaler
func (r UserInfoJwtResponse) MarshalText() (text []byte, err error) {
	return []byte(r), nil
}

type UserInfoEndpoint struct {
	issuer       security.Issuer
	accountStore security.AccountStore
	jwtEncoder   jwt.JwtEncoder
}

func NewUserInfoEndpoint(issuer security.Issuer, accountStore security.AccountStore, jwtEncoder jwt.JwtEncoder) *UserInfoEndpoint {
	return &UserInfoEndpoint{
		issuer:       issuer,
		accountStore: accountStore,
		jwtEncoder:   jwtEncoder,
	}
}

func (ep *UserInfoEndpoint) PlainUserInfo(ctx context.Context, _ UserInfoRequest) (resp *UserInfoPlainResponse, err error) {
	auth, ok := security.Get(ctx).(oauth2.Authentication)
	if !ok || auth.UserAuthentication() == nil {
		return nil, oauth2.NewAccessRejectedError("missing user authentication")
	}

	specs := ep.determineClaimSpecs(auth.OAuth2Request())
	requested := ep.determineRequestedClaims(auth.OAuth2Request())
	c := UserInfoClaims{}
	e := claims.Populate(ctx, &c,
		claims.WithSpecs(specs...),
		claims.WithSource(auth),
		claims.WithIssuer(ep.issuer),
		claims.WithAccountStore(ep.accountStore),
		claims.WithRequestedClaims(requested, fullSpecs...),
	)

	if e != nil {
		return nil, oauth2.NewInternalError(e)
	}

	return &UserInfoPlainResponse{
		UserInfoClaims: c,
	}, nil
}

func (ep *UserInfoEndpoint) JwtUserInfo(ctx context.Context, _ UserInfoRequest) (resp UserInfoJwtResponse, err error) {
	auth, ok := security.Get(ctx).(oauth2.Authentication)
	if !ok || auth.UserAuthentication() == nil {
		return "", oauth2.NewAccessRejectedError("missing user authentication")
	}

	c := UserInfoClaims{}

	e := claims.Populate(ctx, &c,
		claims.WithSpecs(
			claims.UserInfoBasicSpecs,
			claims.ProfileScopeSpecs,
			claims.EmailScopeSpecs,
			claims.PhoneScopeSpecs,
			claims.AddressScopeSpecs,
		),
		claims.WithSource(auth),
		claims.WithIssuer(ep.issuer),
		claims.WithAccountStore(ep.accountStore),
	)
	if e != nil {
		return "", oauth2.NewInternalError(err)
	}

	token, e := ep.jwtEncoder.Encode(ctx, &c)
	if e != nil {
		return "", oauth2.NewInternalError(e)
	}
	return UserInfoJwtResponse(token), nil
}

// determineClaimSpecs works slightly different from the id_token version:
// When openid scope is not in the request, full specs is given
func (ep *UserInfoEndpoint) determineClaimSpecs(request oauth2.OAuth2Request) []map[string]claims.ClaimSpec {
	if request == nil || request.Scopes() == nil || !request.Approved() {
		return defaultSpecs
	}

	if !request.Scopes().Has(oauth2.ScopeOidc) {
		return fullSpecs
	}

	specs := make([]map[string]claims.ClaimSpec, len(defaultSpecs), len(defaultSpecs)+len(request.Scopes()))
	for i, spec := range defaultSpecs {
		specs[i] = spec
	}

	scopes := request.Scopes()
	for scope, spec := range scopedSpecs {
		if scopes.Has(scope) {
			specs = append(specs, spec)
		}
	}
	return specs
}

func (ep *UserInfoEndpoint) determineRequestedClaims(request oauth2.OAuth2Request) claims.RequestedClaims {
	raw, ok := request.Extensions()[oauth2.ParameterClaims].(string)
	if !ok {
		return nil
	}

	cr := openid.ClaimsRequest{}
	if e := json.Unmarshal([]byte(raw), &cr); e != nil {
		return nil
	}
	return cr.UserInfo
}

func JwtResponseEncoder() httptransport.EncodeResponseFunc {
	return web.CustomResponseEncoder(func(opt *web.EncodeOption) {
		opt.ContentType = "application/jwt; charset=utf-8"
		opt.WriteFunc = web.TextWriteFunc
	})
}
