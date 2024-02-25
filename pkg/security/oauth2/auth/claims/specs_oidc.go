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

import "github.com/cisco-open/go-lanai/pkg/security/oauth2"

var (
	ProfileScopeSpecs = map[string]ClaimSpec{
		// Profile Scope
		oauth2.ClaimFullName:          Optional(FullName),
		oauth2.ClaimFirstName:         Optional(FirstName),
		oauth2.ClaimLastName:          Optional(LastName),
		oauth2.ClaimMiddleName:        Unsupported(),
		oauth2.ClaimNickname:          Unsupported(),
		oauth2.ClaimPreferredUsername: Optional(Username),
		oauth2.ClaimProfileUrl:        Unsupported(),
		oauth2.ClaimPictureUrl:        Unsupported(),
		oauth2.ClaimWebsite:           Unsupported(),
		oauth2.ClaimGender:            Unsupported(),
		oauth2.ClaimBirthday:          Unsupported(),
		oauth2.ClaimZoneInfo:          Optional(ZoneInfo),
		oauth2.ClaimLocale:            Optional(Locale),
		oauth2.ClaimCurrency:          Optional(Currency),
		oauth2.ClaimUpdatedAt:         Unsupported(),
		oauth2.ClaimDefaultTenantId:   Optional(DefaultTenantId),
		oauth2.ClaimRoles:             Optional(Roles),
		oauth2.ClaimPermissions:       Optional(Permissions),
	}

	EmailScopeSpecs = map[string]ClaimSpec{
		oauth2.ClaimEmail:         Optional(Email),
		oauth2.ClaimEmailVerified: Optional(EmailVerified),
	}

	PhoneScopeSpecs = map[string]ClaimSpec{
		oauth2.ClaimPhoneNumber:      Unsupported(),
		oauth2.ClaimPhoneNumVerified: Unsupported(),
	}

	AddressScopeSpecs = map[string]ClaimSpec{
		oauth2.ClaimAddress: Optional(Address),
	}
)
