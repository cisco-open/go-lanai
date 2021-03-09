package misc

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
)

type UserInfoRequest struct {}

type UserInfoPlainResponse struct {
	UserInfoClaims
}

type UserInfoEndpoint struct {
	issuer       security.Issuer
	accountStore security.AccountStore
}

func NewUserInfoEndpoint(issuer security.Issuer, accountStore security.AccountStore) *UserInfoEndpoint {
	return &UserInfoEndpoint{
		issuer:       issuer,
		accountStore: accountStore,
	}
}

func (ep *UserInfoEndpoint) PlainUserInfo(ctx context.Context, r UserInfoRequest) (resp *UserInfoPlainResponse, err error) {
	auth, ok := security.Get(ctx).(oauth2.Authentication)
	if !ok || auth.UserAuthentication() == nil {
		return nil, oauth2.NewAccessRejectedError("missing user authentication")
	}

	c := UserInfoClaims{}

	if e := claims.Populate(ctx, &c, claims.UserInfoClaimSpecs,
		claims.WithSource(auth), claims.WithIssuer(ep.issuer), claims.WithAccountStore(ep.accountStore),
	); e != nil {
		return nil, oauth2.NewInternalError(err.Error(), err)
	}

	return &UserInfoPlainResponse{
		UserInfoClaims: c,
	}, nil
}



