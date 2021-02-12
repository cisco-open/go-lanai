package misc

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
)

type CheckTokenRequest struct {
	Value     string `form:"token" binding:"required"`
	Hint      string `form:"token_type_hint"`
	NoDetails bool   `form:"no_details"`
}

type CheckTokenEndpoint struct {
	authenticator tokenauth.Authenticator
}

func NewCheckTokenEndpoint(authenticator tokenauth.Authenticator) *CheckTokenEndpoint {
	return &CheckTokenEndpoint{
		authenticator: authenticator,
	}
}

func (ep *CheckTokenEndpoint) JwkSet(c context.Context, request *CheckTokenRequest) (response *CheckTokenClaims, err error) {
	return nil, nil
}
