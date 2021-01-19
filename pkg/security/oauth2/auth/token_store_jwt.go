package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
)

// jwtTokenStore implements TokenStore and delegate oauth2.TokenStoreReader portion to embedded interface
type jwtTokenStore struct {
	oauth2.TokenStoreReader
	detailsStore security.ContextDetailsStore
}

func NewJwtTokenStore(detailsStore security.ContextDetailsStore) TokenStore {
	reader := oauth2.NewJwtTokenStoreReader(detailsStore)
	return &jwtTokenStore{
		TokenStoreReader: reader,
		detailsStore:     detailsStore,
	}
}

func (i *jwtTokenStore) ReusableAccessToken(c context.Context, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	// JWT don't reuse access token
	return nil, nil
}

func (i *jwtTokenStore) SaveAccessToken(c context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, fmt.Errorf("Unsupported token implementation [%T]", token)
	}

	// TODO encode the token, save details, etc.
	return t, nil
}

func (i *jwtTokenStore) SaveRefreshToken(c context.Context, token oauth2.RefreshToken, oauth oauth2.Authentication) (oauth2.RefreshToken, error) {
	t, ok := token.(*oauth2.DefaultRefreshToken)
	if !ok {
		return nil, fmt.Errorf("Unsupported token implementation [%T]", token)
	}

	// TODO encode the token, save details, etc.
	return t, nil
}

func (i *jwtTokenStore) RemoveAccessToken(c context.Context, token oauth2.Token) error {
	// TODO do the magic
	return nil
}

func (i *jwtTokenStore) RemoveRefreshToken(c context.Context, token oauth2.RefreshToken) error {
	// TODO do the magic
	return nil
}

