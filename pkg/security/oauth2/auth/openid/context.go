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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

var logger = log.New("OpenID")

//goland:noinspection GoUnusedConst
const (
	PromptNone  = `none`
	PromptLogin = `login`
	//PromptConsent    = `consent`
	//PromptSelectAcct = `select_account`
)

const (
	DisplayPage = `page`
	PromptTouch = `touch`
	//PromptPopup = `popup`
	//PromptWap   = `wap`
)

const (
	WellKnownEndpointOPConfig = `/.well-known/openid-configuration`
)

var (
	SupportedGrantTypes = utils.NewStringSet(
		oauth2.GrantTypeAuthCode,
		oauth2.GrantTypeImplicit,
		oauth2.GrantTypePassword,
		oauth2.GrantTypeSwitchUser,
		oauth2.GrantTypeSwitchTenant,
	)
	SupportedDisplayMode  = utils.NewStringSet(DisplayPage, PromptTouch)
	FullIdTokenGrantTypes = utils.NewStringSet(
		oauth2.GrantTypePassword,
		oauth2.GrantTypeSwitchUser,
		oauth2.GrantTypeSwitchTenant,
	)
)

// See https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
//goland:noinspection GoUnusedConst
//nolint:gosec
const (
	OPMetadataIssuer                = "issuer"
	OPMetadataAuthEndpoint          = "authorization_endpoint"
	OPMetadataTokenEndpoint         = "token_endpoint"
	OPMetadataUserInfoEndpoint      = "userinfo_endpoint"
	OPMetadataJwkSetURI             = "jwks_uri"
	OPMetadataRegEndpoint           = "registration_endpoint"
	OPMetadataScopes                = "scopes_supported"
	OPMetadataResponseTypes         = "response_types_supported"
	OPMetadataResponseModes         = "response_modes_supported"
	OPMetadataGrantTypes            = "grant_types_supported"
	OPMetadataACRValues             = "acr_values_supported"
	OPMetadataSubjectTypes          = "subject_types_supported"
	OPMetadataIdTokenJwsAlg         = "id_token_signing_alg_values_supported"
	OPMetadataIdTokenJweAlg         = "id_token_encryption_alg_values_supported"
	OPMetadataIdTokenJweEnc         = "id_token_encryption_enc_values_supported"
	OPMetadataUserInfoJwsAlg        = "userinfo_signing_alg_values_supported"
	OPMetadataUserInfoJweAlg        = "userinfo_encryption_alg_values_supported"
	OPMetadataUserInfoJweEnc        = "userinfo_encryption_enc_values_supported"
	OPMetadataRequestJwsAlg         = "request_object_signing_alg_values_supported"
	OPMetadataRequestJweAlg         = "request_object_encryption_alg_values_supported"
	OPMetadataRequestJweEnc         = "request_object_encryption_enc_values_supported"
	OPMetadataClientAuthMethod      = "token_endpoint_auth_methods_supported"
	OPMetadataAuthJwsAlg            = "token_endpoint_auth_signing_alg_values_supported"
	OPMetadataDisplayValues         = "display_values_supported"
	OPMetadataClaimTypes            = "claim_types_supported"
	OPMetadataClaims                = "claims_supported"
	OPMetadataServiceDocs           = "service_documentation"
	OPMetadataClaimsLocales         = "claims_locales_supported"
	OPMetadataUILocales             = "ui_locales_supported"
	OPMetadataClaimsParams          = "claims_parameter_supported"
	OPMetadataRequestParams         = "request_parameter_supported"
	OPMetadataRequestUriParams      = "request_uri_parameter_supported"
	OPMetadataRequiresRequestUriReg = "require_request_uri_registration"
	OPMetadataPolicyUri             = "op_policy_uri"
	OPMetadataTosUri                = "op_tos_uri"
	OPMetadataEndSessionEndpoint    = "end_session_endpoint"
)

// OPMetadata leverage claims implementations
type OPMetadata struct {
	oauth2.FieldClaimsMapper
	oauth2.MapClaims
	Issuer                 string          `claim:"issuer"`
	AuthEndpoint           string          `claim:"authorization_endpoint"`
	TokenEndpoint          string          `claim:"token_endpoint"`
	UserInfoEndpoint       string          `claim:"userinfo_endpoint"`
	JwkSetURI              string          `claim:"jwks_uri"`
	SupportedGrantTypes    utils.StringSet `claim:"grant_types_supported"`
	SupportedScopes        utils.StringSet `claim:"scopes_supported"`
	SupportedResponseTypes utils.StringSet `claim:"response_types_supported"`
	SupportedACRs          utils.StringSet `claim:"acr_values_supported"`
	SupportedSubjectTypes  utils.StringSet `claim:"subject_types_supported"`
	SupportedIdTokenJwsAlg utils.StringSet `claim:"id_token_signing_alg_values_supported"`
	SupportedClaims        utils.StringSet `claim:"claims_supported"`
}

func (m OPMetadata) MarshalJSON() ([]byte, error) {
	return m.FieldClaimsMapper.DoMarshalJSON(m)
}

func (m *OPMetadata) UnmarshalJSON(bytes []byte) error {
	return m.FieldClaimsMapper.DoUnmarshalJSON(m, bytes)
}

func (m OPMetadata) Get(claim string) interface{} {
	return m.FieldClaimsMapper.Get(m, claim)
}

func (m OPMetadata) Has(claim string) bool {
	return m.FieldClaimsMapper.Has(m, claim)
}

func (m *OPMetadata) Set(claim string, value interface{}) {
	m.FieldClaimsMapper.Set(m, claim, value)
}

func (m OPMetadata) Values() map[string]interface{} {
	return m.FieldClaimsMapper.Values(m)
}
