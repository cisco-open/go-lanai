package grants

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
)

// ClientCredentialsGranter implements auth.TokenGranter
type ClientCredentialsGranter struct {

}

func NewClientCredentialsGranter() *ClientCredentialsGranter {
	return &ClientCredentialsGranter{}
}

func (g *ClientCredentialsGranter) Grant(ctx context.Context, request *auth.TokenRequest) (oauth2.AccessToken, error) {
	//return nil, oauth2.NewInvalidTokenRequestError("invalid token request")
	return oauth2.NewDefaultAccessToken("TODO"), nil
}

