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

package samlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

type DefaultSamlClient struct {
	SamlSpDetails
	TenantRestrictions utils.StringSet
	TenantRestrictionType string
}

func (c DefaultSamlClient) ShouldMetadataRequireSignature() bool {
	return c.MetadataRequireSignature
}

func (c DefaultSamlClient) ShouldMetadataTrustCheck() bool {
	return c.MetadataTrustCheck
}

func (c DefaultSamlClient) GetMetadataTrustedKeys() []string {
	return c.MetadataTrustedKeys
}

func (c DefaultSamlClient) GetEntityId() string {
	return c.EntityId
}

func (c DefaultSamlClient) GetMetadataSource() string {
	return c.MetadataSource
}

func (c DefaultSamlClient) ShouldSkipAssertionEncryption() bool {
	return c.SkipAssertionEncryption
}

func (c DefaultSamlClient) ShouldSkipAuthRequestSignatureVerification() bool {
	return c.SkipAuthRequestSignatureVerification
}

func (c DefaultSamlClient) GetTenantRestrictions() utils.StringSet {
	return c.TenantRestrictions
}

func (c DefaultSamlClient) GetTenantRestrictionType() string {
	return c.TenantRestrictionType
}

type SamlSpDetails struct {
	EntityId string
	MetadataSource string
	SkipAssertionEncryption bool
	SkipAuthRequestSignatureVerification bool

	MetadataRequireSignature bool
	MetadataTrustCheck bool
	MetadataTrustedKeys []string

	//currently the implementation is metaiop profile. this field is reserved for future use
	// https://docs.spring.io/autorepo/docs/spring-security-saml/1.0.x-SNAPSHOT/reference/htmlsingle/#configuration-security-profiles-pkix
	SecurityProfile string
}