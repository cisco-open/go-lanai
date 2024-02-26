// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package auth

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2/common"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2/jwt"
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

func (s *jwtTokenStore) ReadAuthentication(ctx context.Context, tokenValue string, hint oauth2.TokenHint) (oauth2.Authentication, error) {
	switch hint {
	case oauth2.TokenHintRefreshToken:
		return s.readAuthenticationFromRefreshToken(ctx, tokenValue)
	default:
		return s.TokenStoreReader.ReadAuthentication(ctx, tokenValue, hint)
	}
}

func (s *jwtTokenStore) ReusableAccessToken(_ context.Context, _ oauth2.Authentication) (oauth2.AccessToken, error) {
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
	switch t := token.(type) {
	case oauth2.AccessToken:
		// just remove access token
		return s.registry.RevokeAccessToken(c, t)
	case oauth2.RefreshToken:
		// remove all access token associated with this refresh token
		return s.registry.RevokeAllAccessTokens(c, t)
	}
	return nil
}

func (s *jwtTokenStore) RemoveRefreshToken(c context.Context, token oauth2.RefreshToken) error {
	// remove all access token associated with this refresh token and refresh token itself
	return s.registry.RevokeRefreshToken(c, token)
}

/********************
	Helpers
 ********************/
func (s *jwtTokenStore) readAuthenticationFromRefreshToken(c context.Context, tokenValue string) (oauth2.Authentication, error) {
	// parse JWT token
	token, e := s.ReadRefreshToken(c, tokenValue)
	if e != nil {
		return nil, e
	}

	if container, ok := token.(oauth2.ClaimsContainer); !ok || container.Claims() == nil {
		return nil, oauth2.NewInvalidGrantError("refresh token contains no claims")
	}

	stored, e := s.registry.ReadStoredAuthorization(c, token)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError("refresh token unknown", e)
	}

	return stored, nil
}