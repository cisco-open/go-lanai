package testdata

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
)

type CustomAuthRegistry struct {
	RegistrationCount int
}

func NewCustomAuthRegistry() auth.AuthorizationRegistry {
	return &CustomAuthRegistry{}
}

func (c *CustomAuthRegistry) RegisterRefreshToken(ctx context.Context, token oauth2.RefreshToken, oauth oauth2.Authentication) error {
	panic("implement me")
}

func (c *CustomAuthRegistry) RegisterAccessToken(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) error {
	c.RegistrationCount++
	return nil
}

func (c *CustomAuthRegistry) ReadStoredAuthorization(ctx context.Context, token oauth2.RefreshToken) (oauth2.Authentication, error) {
	panic("implement me")
}

func (c *CustomAuthRegistry) FindSessionId(ctx context.Context, token oauth2.Token) (string, error) {
	panic("implement me")
}

func (c *CustomAuthRegistry) RevokeRefreshToken(ctx context.Context, token oauth2.RefreshToken) error {
	panic("implement me")
}

func (c *CustomAuthRegistry) RevokeAccessToken(ctx context.Context, token oauth2.AccessToken) error {
	panic("implement me")
}

func (c *CustomAuthRegistry) RevokeAllAccessTokens(ctx context.Context, token oauth2.RefreshToken) error {
	panic("implement me")
}

func (c *CustomAuthRegistry) RevokeUserAccess(ctx context.Context, username string, revokeRefreshToken bool) error {
	panic("implement me")
}

func (c *CustomAuthRegistry) RevokeClientAccess(ctx context.Context, clientId string, revokeRefreshToken bool) error {
	panic("implement me")
}

func (c CustomAuthRegistry) RevokeSessionAccess(ctx context.Context, sessionId string, revokeRefreshToken bool) error {
	panic("implement me")
}
