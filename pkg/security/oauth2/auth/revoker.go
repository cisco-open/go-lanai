package auth

import (
	"context"
)

const (
	RevokerHintAccessToken  RevokerTokenHint = "access_token"
	RevokerHintRefreshToken RevokerTokenHint = "refresh_token"
)

type RevokerTokenHint string

//go:generate mockery --name AccessRevoker
type AccessRevoker interface {
	RevokeWithSessionId(ctx context.Context, sessionId string, sessionName string) error
	RevokeWithUsername(ctx context.Context, username string, revokeRefreshToken bool) error
	RevokeWithClientId(ctx context.Context, clientId string, revokeRefreshToken bool) error
	RevokeWithTokenValue(ctx context.Context, tokenValue string, hint RevokerTokenHint) error
}
