package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

type TokenStore interface {
	oauth2.TokenStoreReader

	// ReusableAccessToken finds access token that currently associated with given oauth2.Authentication
	// and can be reused
	ReusableAccessToken(ctx context.Context, oauth oauth2.Authentication) (oauth2.AccessToken, error)

	// SaveAccessToken associate given oauth2.Authentication with the to-be-saved oauth2.AccessToken.
	// It returns the saved oauth2.AccessToken or error.
	// The saved oauth2.AccessToken may be different from given oauth2.AccessToken (e.g. JWT encoded token)
	SaveAccessToken(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error)

	// SaveRefreshToken associate given oauth2.Authentication with the to-be-saved oauth2.RefreshToken.
	// It returns the saved oauth2.RefreshToken or error.
	// The saved oauth2.RefreshToken may be different from given oauth2.RefreshToken (e.g. JWT encoded token)
	SaveRefreshToken(ctx context.Context, token oauth2.RefreshToken, oauth oauth2.Authentication) (oauth2.RefreshToken, error)

	// RemoveAccessToken remove oauth2.AccessToken using given token value.
	// Token can be oauth2.AccessToken or oauth2.RefreshToken
	RemoveAccessToken(ctx context.Context, token oauth2.Token) error

	// RemoveRefreshToken remove given oauth2.RefreshToken
	RemoveRefreshToken(ctx context.Context, token oauth2.RefreshToken) error
}
