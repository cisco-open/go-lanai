package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

type AccessRevoker interface {
	RevokeWithSessionId(ctx context.Context, sessionId string, sessionName string) error
	RevokeWithUsername(ctx context.Context, username string, revokeRefreshToken bool) error
	RevokeWithClientId(ctx context.Context, clientId string, revokeRefreshToken bool) error
	RevokeRefreshToken(ctx context.Context, token oauth2.RefreshToken) error
	RevokeAccessToken(ctx context.Context, token oauth2.AccessToken) error
	RevokeAllAccessTokens(ctx context.Context, token oauth2.RefreshToken) error
}
