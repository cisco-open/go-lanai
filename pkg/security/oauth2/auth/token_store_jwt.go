package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"fmt"
)

// jwtTokenStore implements TokenStore and delegate oauth2.TokenStoreReader portion to embedded interface
type jwtTokenStore struct {
	oauth2.TokenStoreReader
	detailsStore security.ContextDetailsStore
	jwtEncoder jwt.JwtEncoder
}

func NewJwtTokenStore(reader oauth2.TokenStoreReader, detailsStore security.ContextDetailsStore, jwtEncoder jwt.JwtEncoder) TokenStore {
	return &jwtTokenStore{
		TokenStoreReader: reader,
		detailsStore:     detailsStore,
		jwtEncoder:       jwtEncoder,
	}
}

func (s *jwtTokenStore) ReusableAccessToken(c context.Context, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	// JWT don't reuse access token
	return nil, nil
}

func (s *jwtTokenStore) SaveAccessToken(c context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, fmt.Errorf("Unsupported token implementation [%T]", token)
	} else if t.Claims == nil {
		return nil, fmt.Errorf("claims is nil")
	}

	encoded, e := s.jwtEncoder.Encode(c, t.Claims)
	if e != nil {
		return nil, e
	}
	t.SetValue(encoded)

	// TODO save details, etc.
	return t, nil
}

func (s *jwtTokenStore) SaveRefreshToken(c context.Context, token oauth2.RefreshToken, oauth oauth2.Authentication) (oauth2.RefreshToken, error) {
	t, ok := token.(*oauth2.DefaultRefreshToken)
	if !ok {
		return nil, fmt.Errorf("Unsupported token implementation [%T]", token)
	} else if t.Claims == nil {
		return nil, fmt.Errorf("claims is nil")
	}

	encoded, e := s.jwtEncoder.Encode(c, t.Claims)
	if e != nil {
		return nil, e
	}
	t.SetValue(encoded)

	// TODO save details, etc.
	return t, nil
}

func (s *jwtTokenStore) RemoveAccessToken(c context.Context, token oauth2.Token) error {
	// TODO do the magic
	return nil
}

func (s *jwtTokenStore) RemoveRefreshToken(c context.Context, token oauth2.RefreshToken) error {
	// TODO do the magic
	return nil
}

