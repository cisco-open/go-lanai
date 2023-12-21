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

	// ReadAccessToken load AccessToken with given value.
	// If the AccessToken is not associated with a valid security.ContextDetails (revoked), it returns error
	ReadAccessToken(ctx context.Context, value string) (AccessToken, error)

	// ReadRefreshToken load RefreshToken with given value.
	// this method does not imply any revocation status. it depends on implementation
	ReadRefreshToken(ctx context.Context, value string) (RefreshToken, error)
}




