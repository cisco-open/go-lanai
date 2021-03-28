package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

// AuthorizationRegistry is responsible to keep track of refresh token and relationships between tokens, clients, users, sessions
type AuthorizationRegistry interface {
	// Register
	RegisterRefreshToken(ctx context.Context, token oauth2.RefreshToken, oauth oauth2.Authentication) error
	RegisterAccessToken(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) error

	// Read
	ReadStoredAuthorization(ctx context.Context, token oauth2.RefreshToken) (oauth2.Authentication, error)
	RefreshTokenExists(ctx context.Context, token oauth2.RefreshToken) bool
	FindSessionId(ctx context.Context, token oauth2.Token) (string, error)

	// Revoke
	RevokeRefreshToken(ctx context.Context, token oauth2.RefreshToken) error
	RevokeAccessToken(ctx context.Context, token oauth2.AccessToken) error
	RevokeAllAccessTokens(ctx context.Context, token oauth2.RefreshToken) error
	RevokeUserAccess(ctx context.Context, username string) error
	RevokeClientAccess(ctx context.Context, clientId string) error
}
