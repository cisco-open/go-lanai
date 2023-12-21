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

package extsamlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
)

type SamlIdpAutoCreateUserDetails struct {
	Enabled bool
	EmailWhiteList []string
	AttributeMapping map[string]string
	ElevatedUserRoleNames []string
	RegularUserRoleNames []string
}

func (a SamlIdpAutoCreateUserDetails) GetElevatedUserRoleNames() []string {
	return a.ElevatedUserRoleNames
}

func (a SamlIdpAutoCreateUserDetails) GetRegularUserRoleNames() []string {
	return a.RegularUserRoleNames
}

func (a SamlIdpAutoCreateUserDetails) IsEnabled() bool {
	return a.Enabled
}

func (a SamlIdpAutoCreateUserDetails) GetEmailWhiteList() []string {
	return a.EmailWhiteList
}

func (a SamlIdpAutoCreateUserDetails) GetAttributeMapping() map[string]string {
	return a.AttributeMapping
}

type SamlIdpDetails struct {
	EntityId                 string
	Domain                   string
	MetadataLocation         string
	ExternalIdName           string
	ExternalIdpName          string
	MetadataRequireSignature bool
	MetadataTrustCheck       bool
	MetadataTrustedKeys      []string
	AutoCreateUserDetails    SamlIdpAutoCreateUserDetails
}


type SamlIdpOptions func(opt *SamlIdpDetails)

// SamlIdentityProvider implements idp.IdentityProvider, idp.AuthenticationFlowAware and samllogin.SamlIdentityProvider
type SamlIdentityProvider struct {
	SamlIdpDetails
}

func (s SamlIdentityProvider) ShouldMetadataRequireSignature() bool {
	return s.MetadataRequireSignature
}

func (s SamlIdentityProvider) ShouldMetadataTrustCheck() bool {
	return s.MetadataTrustCheck
}

func (s SamlIdentityProvider) GetMetadataTrustedKeys() []string {
	return s.MetadataTrustedKeys
}

func NewIdentityProvider(opts ...SamlIdpOptions) *SamlIdentityProvider {
	opt := SamlIdpDetails{}
	for _, f := range opts {
		f(&opt)
	}
	return &SamlIdentityProvider{
		SamlIdpDetails: opt,
	}
}

func (s SamlIdentityProvider) AuthenticationFlow() idp.AuthenticationFlow {
	return idp.ExternalIdpSAML
}

func (s SamlIdentityProvider) Domain() string {
	return s.SamlIdpDetails.Domain
}

func (s SamlIdentityProvider) EntityId() string {
	return s.SamlIdpDetails.EntityId
}

func (s SamlIdentityProvider) MetadataLocation() string {
	return s.SamlIdpDetails.MetadataLocation
}

func (s SamlIdentityProvider) ExternalIdName() string {
	return s.SamlIdpDetails.ExternalIdName
}

func (s SamlIdentityProvider) ExternalIdpName() string {
	return s.SamlIdpDetails.ExternalIdpName
}

func (s SamlIdentityProvider) GetAutoCreateUserDetails() security.AutoCreateUserDetails {
	return s.SamlIdpDetails.AutoCreateUserDetails
}