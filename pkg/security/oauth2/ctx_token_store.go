package oauth2

import "context"

type TokenStoreReader interface {
	// ReadAuthentication load associated Authentication with Token.
	// Token can be AccessToken or RefreshToken
	ReadAuthentication(ctx context.Context, token Token) (Authentication, error)

	// ReadAccessToken load AccessToken with given value
	ReadAccessToken(ctx context.Context, value string) (AccessToken, error)

	// ReadRefreshToken load RefreshToken with given Value
	ReadRefreshToken(ctx context.Context, value string) (RefreshToken, error)
}




