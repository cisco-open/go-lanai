package misc

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	httptransport "github.com/go-kit/kit/transport/http"
)

type UserInfoRequest struct {}

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

	c := UserInfoClaims{}

	if e := claims.Populate(ctx, &c, claims.UserInfoClaimSpecs,
		claims.WithSource(auth), claims.WithIssuer(ep.issuer), claims.WithAccountStore(ep.accountStore),
	); e != nil {
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

	if e := claims.Populate(ctx, &c, claims.UserInfoClaimSpecs,
		claims.WithSource(auth), claims.WithIssuer(ep.issuer), claims.WithAccountStore(ep.accountStore),
	); e != nil {
		return "", oauth2.NewInternalError(err)
	}

	token, e := ep.jwtEncoder.Encode(ctx, &c)
	if e != nil {
		return "", oauth2.NewInternalError(e)
	}
	return UserInfoJwtResponse(token), nil
}

func JwtResponseEncoder() httptransport.EncodeResponseFunc {
	return web.CustomResponseEncoder(func(opt *web.EncodeOption) {
		opt.ContentType = "application/jwt; charset=utf-8"
		opt.WriteFunc = web.TextWriteFunc
	})
}
