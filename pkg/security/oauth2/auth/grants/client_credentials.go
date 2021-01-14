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
	if oauth2.GrantTypeClientCredentials != request.GrantType {
		return nil, nil
	}

	client := auth.RetrieveAuthenticatedClient(ctx)

	// check scope
	if e := auth.ValidateScope(ctx, request, client); e != nil {
		return nil, e
	}

	// TODO create real token
	return oauth2.NewDefaultAccessToken("TODO"), nil
}

