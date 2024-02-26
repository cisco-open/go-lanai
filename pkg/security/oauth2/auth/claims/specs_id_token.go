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

package claims

import (
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
)

var (
	IdTokenBasicSpecs = map[string]ClaimSpec{
		// Basic
		oauth2.ClaimIssuer:          Required(Issuer),
		oauth2.ClaimSubject:         Required(Subject),
		oauth2.ClaimAudience:        Required(Audience),
		oauth2.ClaimExpire:          Required(ExpiresAt),
		oauth2.ClaimIssueAt:         Required(IssuedAt),
		oauth2.ClaimAuthTime:        RequiredIfParamsExists(AuthenticationTime, oauth2.ParameterMaxAge),
		oauth2.ClaimNonce:           RequiredIfParamsExists(Nonce, oauth2.ParameterNonce),
		oauth2.ClaimAuthCtxClassRef: Optional(AuthContextClassRef),
		oauth2.ClaimAuthMethodRef:   Optional(AuthMethodRef),
		oauth2.ClaimAuthorizedParty: Optional(ClientId),
		oauth2.ClaimAccessTokenHash: RequiredIfImplicitFlow(AccessTokenHash),

		// Custom Profile
		oauth2.ClaimUserId:           Optional(UserId),
		oauth2.ClaimAccountType:      Optional(AccountType),
		oauth2.ClaimTenantId:         Optional(TenantId),
		oauth2.ClaimTenantExternalId: Optional(TenantExternalId),
		oauth2.ClaimTenantSuspended:  Optional(TenantSuspended),
		oauth2.ClaimProviderId:       Optional(ProviderId),
		oauth2.ClaimProviderName:     Optional(ProviderName),
		oauth2.ClaimOrigUsername:     Optional(OriginalUsername),
		oauth2.ClaimRoles:            Optional(Roles),
	}
)
