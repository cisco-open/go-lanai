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

package openid

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
)

var (
	OPMetadataBasicSpecs = map[string]claims.ClaimSpec{
		OPMetadataIssuer:           claims.Optional(claims.Issuer),
		OPMetadataAuthEndpoint:     opMetaEndpoint(OPMetadataAuthEndpoint),
		OPMetadataTokenEndpoint:    opMetaEndpoint(OPMetadataTokenEndpoint),
		OPMetadataUserInfoEndpoint: opMetaEndpoint(OPMetadataUserInfoEndpoint),
		OPMetadataJwkSetURI:        opMetaEndpoint(OPMetadataJwkSetURI),
		OPMetadataGrantTypes: opMetaFixedSet(
			oauth2.GrantTypeClientCredentials, oauth2.GrantTypePassword,
			oauth2.GrantTypeAuthCode, oauth2.GrantTypeImplicit, oauth2.GrantTypeRefresh,
			oauth2.GrantTypeSwitchUser, oauth2.GrantTypeSwitchTenant, oauth2.GrantTypeSamlSSO,
		),
		OPMetadataScopes: opMetaFixedSet(
			oauth2.ScopeRead, oauth2.ScopeWrite, oauth2.ScopeTokenDetails, oauth2.ScopeTenantHierarchy,
			oauth2.ScopeOidc, oauth2.ScopeOidcProfile, oauth2.ScopeOidcEmail, oauth2.ScopeOidcAddress, oauth2.ScopeOidcPhone,
		),
		// TODO Add implicit flow support
		//OPMetadataResponseTypes: opMetaFixedSet("code", "id_token", "token id_token"),
		OPMetadataResponseTypes: opMetaFixedSet("code"),
		OPMetadataACRValues:     opMetaAcrValues(1, 2, 3),
		OPMetadataSubjectTypes:  opMetaFixedSet("public"),
		OPMetadataIdTokenJwsAlg: opMetaFixedSet("RS256"),
		OPMetadataClaims: opMetaFixedSet(
			oauth2.ClaimIssuer, oauth2.ClaimSubject, oauth2.ClaimAudience, oauth2.ClaimExpire, oauth2.ClaimIssueAt,
			oauth2.ClaimAuthTime, oauth2.ClaimNonce, oauth2.ClaimAuthCtxClassRef, oauth2.ClaimAuthMethodRef, oauth2.ClaimAuthorizedParty,
			oauth2.ClaimFullName, oauth2.ClaimFirstName, oauth2.ClaimLastName, oauth2.ClaimPreferredUsername,
			oauth2.ClaimEmail, oauth2.ClaimEmailVerified, oauth2.ClaimLocale,
			//oauth2.ClaimPhoneNumber, oauth2.ClaimPhoneNumVerified, oauth2.ClaimAddress,
		),
	}

	OPMetadataOptionalSpecs = map[string]claims.ClaimSpec{
		OPMetadataRegEndpoint:           claims.Unsupported(),
		OPMetadataResponseModes:         claims.Unsupported(),
		OPMetadataIdTokenJweAlg:         claims.Unsupported(),
		OPMetadataIdTokenJweEnc:         claims.Unsupported(),
		OPMetadataUserInfoJwsAlg:        opMetaFixedSet("RS256"),
		OPMetadataUserInfoJweAlg:        claims.Unsupported(),
		OPMetadataUserInfoJweEnc:        claims.Unsupported(),
		OPMetadataRequestJwsAlg:         claims.Unsupported(),
		OPMetadataRequestJweAlg:         claims.Unsupported(),
		OPMetadataRequestJweEnc:         claims.Unsupported(),
		OPMetadataClientAuthMethod:      opMetaFixedSet("client_secret_basic", "client_secret_post"),
		OPMetadataAuthJwsAlg:            claims.Unsupported(),
		OPMetadataDisplayValues:         opMetaFixedSet("page", "touch"),
		OPMetadataClaimTypes:            opMetaFixedSet("normal"),
		OPMetadataServiceDocs:           claims.Unsupported(),
		OPMetadataClaimsLocales:         opMetaFixedSet("en-CA", "en-US"),
		OPMetadataUILocales:             opMetaFixedSet("en-CA", "en-US"),
		OPMetadataClaimsParams:          opMetaFixedBool(true),
		OPMetadataRequestParams:         opMetaFixedBool(true),
		OPMetadataRequestUriParams:      claims.Unsupported(),
		OPMetadataRequiresRequestUriReg: claims.Unsupported(),
		OPMetadataPolicyUri:             claims.Unsupported(),
		OPMetadataTosUri:                claims.Unsupported(),
		OPMetadataEndSessionEndpoint:    opMetaEndpoint(OPMetadataEndSessionEndpoint),
	}
)
