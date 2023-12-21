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
)

const (
	RevokerHintAccessToken  RevokerTokenHint = "access_token"
	RevokerHintRefreshToken RevokerTokenHint = "refresh_token"
)

type RevokerTokenHint string

//go:generate mockery --name AccessRevoker
type AccessRevoker interface {
	RevokeWithSessionId(ctx context.Context, sessionId string, sessionName string) error
	RevokeWithUsername(ctx context.Context, username string, revokeRefreshToken bool) error
	RevokeWithClientId(ctx context.Context, clientId string, revokeRefreshToken bool) error
	RevokeWithTokenValue(ctx context.Context, tokenValue string, hint RevokerTokenHint) error
}
