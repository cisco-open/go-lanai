package grants

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"fmt"
)

// ClientCredentialsGranter implements auth.TokenGranter
type PasswordGranter struct {
	authenticator security.Authenticator
	authService   auth.AuthorizationService
}

func NewPasswordGranter(authenticator security.Authenticator, authService auth.AuthorizationService) *PasswordGranter {
	if authenticator == nil {
		panic(fmt.Errorf("cannot create PasswordGranter without authenticator."))
	}

	if authService == nil {
		panic(fmt.Errorf("cannot create PasswordGranter without token service."))
	}

	return &PasswordGranter{
		authenticator: authenticator,
		authService:   authService,
	}
}

func (g *PasswordGranter) Grant(ctx context.Context, request *auth.TokenRequest) (oauth2.AccessToken, error) {
	if oauth2.GrantTypePassword != request.GrantType {
		return nil, nil
	}

	client := auth.RetrieveAuthenticatedClient(ctx)

	// common check
	if e := CommonPreGrantValidation(ctx, client, request); e != nil {
		return nil, e
	}

	// extract username & password
	username, uOk := request.Parameters[oauth2.ParameterUsername]
	password, pOk := request.Parameters[oauth2.ParameterPassword]
	delete(request.Parameters, oauth2.ParameterPassword)
	if !uOk || !pOk {
		return nil, oauth2.NewInvalidTokenRequestError("missing 'username' and 'password'")
	}

	// authenticate
	candidate := passwd.UsernamePasswordPair{
		Username: username,
		Password: password,
	}

	userAuth, err := g.authenticator.Authenticate(ctx, &candidate)
	if err != nil || userAuth.State() < security.StateAuthenticated {
		return nil, oauth2.NewInvalidTokenRequestError(err.Error(), err)
	}

	// additional check
	if request.Scopes == nil || len(request.Scopes) == 0 {
		request.Scopes = client.Scopes()
	}

	if e := auth.ValidateAllAutoApprovalScopes(ctx, client, request.Scopes); e != nil {
		return nil, e
	}

	// create authentication
	req := request.OAuth2Request(client)
	oauth, e := g.authService.CreateAuthentication(ctx, req, userAuth)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(e.Error(), e)
	}

	// create token
	token, e := g.authService.CreateAccessToken(ctx, oauth)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(e.Error(), e)
	}
	return token, nil
}


