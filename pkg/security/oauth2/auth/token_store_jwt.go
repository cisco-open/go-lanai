package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"fmt"
)

// jwtTokenStore implements TokenStore and delegate oauth2.TokenStoreReader portion to embedded interface
type jwtTokenStore struct {
	oauth2.TokenStoreReader
	detailsStore security.ContextDetailsStore
	jwtEncoder   jwt.JwtEncoder
	registry     AuthorizationRegistry
}

type JTSOptions func(opt *JTSOption)

type JTSOption struct {
	Reader       oauth2.TokenStoreReader
	DetailsStore security.ContextDetailsStore
	Encoder      jwt.JwtEncoder
	Decoder      jwt.JwtDecoder
	AuthRegistry AuthorizationRegistry
}

func NewJwtTokenStore(opts...JTSOptions) *jwtTokenStore {
	opt := JTSOption{}
	for _, optFunc := range opts {
		optFunc(&opt)
	}

	if opt.Reader == nil {
		opt.Reader = common.NewJwtTokenStoreReader(func(o *common.JTSROption) {
			o.DetailsStore = opt.DetailsStore
			o.Decoder = opt.Decoder
		})
	}
	return &jwtTokenStore{
		TokenStoreReader: opt.Reader,
		detailsStore:     opt.DetailsStore,
		jwtEncoder:       opt.Encoder,
		registry:         opt.AuthRegistry,
	}
}

func (r *jwtTokenStore) ReadAuthentication(ctx context.Context, tokenValue string, hint oauth2.TokenHint) (oauth2.Authentication, error) {
	switch hint {
	case oauth2.TokenHintRefreshToken:
		return r.readAuthenticationFromRefreshToken(ctx, tokenValue)
	default:
		return r.TokenStoreReader.ReadAuthentication(ctx, tokenValue, hint)
	}
}

func (s *jwtTokenStore) ReusableAccessToken(c context.Context, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	// JWT don't reuse access token
	return nil, nil
}

func (s *jwtTokenStore) SaveAccessToken(c context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, oauth2.NewInternalError(fmt.Sprintf("Unsupported token implementation [%T]", token))
	} else if t.Claims() == nil {
		return nil, oauth2.NewInternalError("claims is nil")
	}

	encoded, e := s.jwtEncoder.Encode(c, t.Claims())
	if e != nil {
		return nil, e
	}
	t.SetValue(encoded)

	if details, ok := oauth.Details().(security.ContextDetails); ok {
		if e := s.detailsStore.SaveContextDetails(c, token, details); e != nil {
			return nil, oauth2.NewInternalError("cannot save access token", e)
		}
	}

	if e := s.registry.RegisterAccessToken(c, t, oauth); e != nil {
		return nil, oauth2.NewInternalError("cannot register access token", e)
	}
	return t, nil
}

func (s *jwtTokenStore) SaveRefreshToken(c context.Context, token oauth2.RefreshToken, oauth oauth2.Authentication) (oauth2.RefreshToken, error) {
	t, ok := token.(*oauth2.DefaultRefreshToken)
	if !ok {
		return nil, fmt.Errorf("Unsupported token implementation [%T]", token)
	} else if t.Claims() == nil {
		return nil, fmt.Errorf("claims is nil")
	}

	encoded, e := s.jwtEncoder.Encode(c, t.Claims())
	if e != nil {
		return nil, e
	}
	t.SetValue(encoded)

	if e := s.registry.RegisterRefreshToken(c, t, oauth); e != nil {
		return nil, oauth2.NewInternalError("cannot register refresh token", e)
	}
	return t, nil
}

func (s *jwtTokenStore) RemoveAccessToken(c context.Context, token oauth2.Token) error {
	switch token.(type) {
	case oauth2.AccessToken:
		// TODO just remove access token
	case oauth2.RefreshToken:
		// TODO remove all access token associated with this refresh token
	}
	return nil
}

func (s *jwtTokenStore) RemoveRefreshToken(c context.Context, token oauth2.RefreshToken) error {
	// TODO remove all access token associated with this refresh token and refresh token itself
	return nil
}

/********************
	Helpers
 ********************/
func (r *jwtTokenStore) readAuthenticationFromRefreshToken(c context.Context, tokenValue string) (oauth2.Authentication, error) {
	// parse JWT token
	token, e := r.ReadRefreshToken(c, tokenValue)
	if e != nil {
		return nil, e
	}

	if container, ok := token.(oauth2.ClaimsContainer); !ok || container.Claims() == nil {
		return nil, oauth2.NewInvalidGrantError("refresh token contains no claims")
	}

	stored, e := r.registry.ReadStoredAuthorization(c, token)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError("refresh token unknown", e)
	}

	return stored, nil
}