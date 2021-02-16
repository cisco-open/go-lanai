package oauth2

import "context"

const (
	_ TokenHint = iota
	TokenHintAccessToken
	TokenHintRefreshToken
)

type TokenHint int

func (h TokenHint) String() string {
	switch h {
	case TokenHintAccessToken:
		return "access_token"
	case TokenHintRefreshToken:
		return "refresh_token"
	default:
		return "unknown"
	}
}

type TokenStoreReader interface {
	// ReadAuthentication load associated Authentication with Token.
	// Token can be AccessToken or RefreshToken
	ReadAuthentication(ctx context.Context, tokenValue string, hint TokenHint) (Authentication, error)

	// ReadAccessToken load AccessToken with given value
	ReadAccessToken(ctx context.Context, value string) (AccessToken, error)

	// ReadRefreshToken load RefreshToken with given Token
	ReadRefreshToken(ctx context.Context, value string) (RefreshToken, error)
}




