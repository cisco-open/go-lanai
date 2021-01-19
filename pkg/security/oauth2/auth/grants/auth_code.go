package grants

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"fmt"
)

// ClientCredentialsGranter implements auth.TokenGranter
type AuthorizationCodeGranter struct {
	tokenService auth.AuthorizationService
}

func NewAuthorizationCodeGranter(tokenService auth.AuthorizationService) *AuthorizationCodeGranter {
	if tokenService == nil {
		panic(fmt.Errorf("cannot create AuthorizationCodeGranter without token service."))
	}

	return &AuthorizationCodeGranter{
		tokenService: tokenService,
	}
}

func (g *AuthorizationCodeGranter) Grant(ctx context.Context, request *auth.TokenRequest) (oauth2.AccessToken, error) {
	if oauth2.GrantTypeAuthCode != request.GrantType {
		return nil, nil
	}

	client := auth.RetrieveAuthenticatedClient(ctx)

	// common check
	if e := CommonPreGrantValidation(ctx, client, request); e != nil {
		return nil, e
	}

	// TODO create real token
	return oauth2.NewDefaultAccessToken("TODO"), nil
}


