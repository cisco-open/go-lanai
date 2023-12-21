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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

var (
	CheckTokenClaimSpecs = map[string]ClaimSpec{
		// Basic
		oauth2.ClaimAudience:  Required(Audience),
		oauth2.ClaimExpire:    Optional(ExpiresAt),
		oauth2.ClaimJwtId:     Optional(JwtId),
		oauth2.ClaimIssueAt:   Optional(IssuedAt),
		oauth2.ClaimIssuer:    Required(Issuer),
		oauth2.ClaimNotBefore: Optional(NotBefore),
		oauth2.ClaimSubject:   Optional(Subject),
		oauth2.ClaimScope:     Optional(Scopes),
		oauth2.ClaimClientId:  Required(ClientId),
		oauth2.ClaimUsername:  Optional(Username),

		// OIDC
		oauth2.ClaimAuthTime:  Optional(AuthenticationTime),
		oauth2.ClaimFirstName: Optional(FirstName),
		oauth2.ClaimLastName:  Optional(LastName),
		oauth2.ClaimEmail:     Optional(Email),
		oauth2.ClaimLocale:    Optional(Locale),

		// Custom
		oauth2.ClaimUserId:                   Optional(UserId),
		oauth2.ClaimAccountType:              Optional(AccountType),
		oauth2.ClaimCurrency:                 Optional(Currency),
		oauth2.ClaimDefaultTenantId:          Optional(DefaultTenantId),
		oauth2.ClaimTenantId:                 Optional(TenantId),
		oauth2.ClaimTenantExternalId:         Optional(TenantExternalId),
		oauth2.ClaimTenantSuspended:          Optional(TenantSuspended),
		oauth2.ClaimProviderId:               Optional(ProviderId),
		oauth2.ClaimProviderName:             Optional(ProviderName),
		oauth2.ClaimProviderDisplayName:      Optional(ProviderDisplayName),
		oauth2.ClaimProviderDescription:      Optional(ProviderDescription),
		oauth2.ClaimProviderEmail:            Optional(ProviderEmail),
		oauth2.ClaimProviderNotificationType: Optional(ProviderNotificationType),

		oauth2.ClaimRoles:        Optional(Roles),
		oauth2.ClaimPermissions:  Optional(Permissions),
		oauth2.ClaimOrigUsername: Optional(OriginalUsername),
	}
)
