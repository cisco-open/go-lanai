package grants

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"fmt"
)

// ClientCredentialsGranter implements auth.TokenGranter
type ClientCredentialsGranter struct {
	authService auth.AuthorizationService
}

func NewClientCredentialsGranter(authService auth.AuthorizationService) *ClientCredentialsGranter {
	if authService == nil {
		panic(fmt.Errorf("cannot create ClientCredentialsGranter without token service."))
	}

	return &ClientCredentialsGranter{
		authService: authService,
	}
}

func (g *ClientCredentialsGranter) Grant(ctx context.Context, request *auth.TokenRequest) (oauth2.AccessToken, error) {
	if oauth2.GrantTypeClientCredentials != request.GrantType {
		return nil, nil
	}

	client := auth.RetrieveAuthenticatedClient(ctx)

	// common check
	if e := CommonPreGrantValidation(ctx, client, request); e != nil {
		return nil, e
	}

	// create authentication
	req := request.OAuth2Request(client)
	oauth, e := g.authService.CreateAuthentication(ctx, req, nil)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(e.Error(), e)
	}

	return g.authService.CreateAccessToken(ctx, oauth)
}

