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
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
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
