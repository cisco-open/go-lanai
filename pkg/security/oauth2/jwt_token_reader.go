package oauth2

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
)

// jwtTokenStoreReader implements TokenStoreReader
type jwtTokenStoreReader struct {
	detailsStore security.ContextDetailsStore
}

func NewJwtTokenStoreReader(detailsStore security.ContextDetailsStore) TokenStoreReader {
	return &jwtTokenStoreReader{
		detailsStore: detailsStore,
	}
}

func (r *jwtTokenStoreReader) ReadAuthentication(c context.Context, token Token) (Authentication, error) {
	panic("implement me")
}

func (r *jwtTokenStoreReader) ReadAccessToken(c context.Context, value string) (AccessToken, error) {
	token := NewDefaultAccessToken(value)
	// TODO decode JWT
	return token, nil
}

func (r *jwtTokenStoreReader) ReadRefreshToken(c context.Context, value string) (RefreshToken, error) {
	token := NewDefaultRefreshToken(value)
	// TODO decode JWT
	return token, nil
}
