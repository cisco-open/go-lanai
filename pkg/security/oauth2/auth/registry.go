package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

// AuthorizationRegistry is responsible to keep track of refresh token and relationships between tokens, clients, users, sessions
type AuthorizationRegistry interface {
	ReadStoredAuthorization(ctx context.Context, token oauth2.RefreshToken) (oauth2.Authentication, error)
	RegisterRefreshToken(ctx context.Context, token oauth2.RefreshToken, oauth oauth2.Authentication) error
	RefreshTokenExists(ctx context.Context, token oauth2.RefreshToken) bool

	// TODO revoke user, revoke client, revoke refresh token, etc
	RevokeRefreshToken(ctx context.Context, token oauth2.RefreshToken) error
}
