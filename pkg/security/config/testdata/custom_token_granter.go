package testdata

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"time"
)

type CustomContextDetails struct {
}

func (c CustomContextDetails) ExpiryTime() time.Time {
	return time.Now().Add(time.Minute)
}

func (c CustomContextDetails) IssueTime() time.Time {
	return time.Now()
}

func (c CustomContextDetails) Roles() utils.StringSet {
	return utils.NewStringSet()
}

func (c CustomContextDetails) Permissions() utils.StringSet {
	return utils.NewStringSet()
}

func (c CustomContextDetails) AuthenticationTime() time.Time {
	return time.Now()
}

type CustomTokenGranter struct {
	authService auth.AuthorizationService
}

func (c *CustomTokenGranter) Inject(authService auth.AuthorizationService) {
	c.authService = authService
}

func NewCustomTokenGranter() auth.TokenGranter {
	return &CustomTokenGranter{}
}

func (c *CustomTokenGranter) Grant(ctx context.Context, request *auth.TokenRequest) (oauth2.AccessToken, error) {
	if "custom_grant" != request.GrantType {
		return nil, nil
	}

	client, e := auth.RetrieveFullyAuthenticatedClient(ctx)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError("requires client secret validated")
	}

	req := request.OAuth2Request(client)
	oauth := oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
		opt.Request = req
		opt.UserAuth = oauth2.NewUserAuthentication(func(opt *oauth2.UserAuthOption) {
			opt.Principal = "custom_grant_principal"
			opt.State = security.StateAuthenticated
		})
		opt.Details = &CustomContextDetails{}
	})
	return c.authService.CreateAccessToken(ctx, oauth)
}
