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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

// AuthorizationRegistry is responsible to keep track of refresh token and relationships between tokens, clients, users, sessions
type AuthorizationRegistry interface {
	// Register
	RegisterRefreshToken(ctx context.Context, token oauth2.RefreshToken, oauth oauth2.Authentication) error
	RegisterAccessToken(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) error

	// Read
	ReadStoredAuthorization(ctx context.Context, token oauth2.RefreshToken) (oauth2.Authentication, error)
	FindSessionId(ctx context.Context, token oauth2.Token) (string, error)

	// Revoke
	RevokeRefreshToken(ctx context.Context, token oauth2.RefreshToken) error
	RevokeAccessToken(ctx context.Context, token oauth2.AccessToken) error
	RevokeAllAccessTokens(ctx context.Context, token oauth2.RefreshToken) error
	RevokeUserAccess(ctx context.Context, username string, revokeRefreshToken bool) error
	RevokeClientAccess(ctx context.Context, clientId string, revokeRefreshToken bool) error
	RevokeSessionAccess(ctx context.Context, sessionId string, revokeRefreshToken bool) error
}
