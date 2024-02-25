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

package samltest

import "github.com/cisco-open/go-lanai/pkg/security"

type ExtSamlMetadata struct {
	EntityId         string
	Domain           string
	Source           string
	Name             string
	IdName           string
	RequireSignature bool
	TrustCheck       bool
	TrustedKeys      []string
}

func NewMockedIdpProvider(opts ...IDPMockOptions) *MockedIdpProvider {
	defaultEntityID, _ := DefaultIssuer.BuildUrl()
	opt := IDPMockOption{
		Properties: IDPProperties{
			ProviderProperties: ProviderProperties{
				EntityID: defaultEntityID.String(),
			},
			SSOPath: "/sso",
			SLOPath: "/slo",
		},
	}
	for _, fn := range opts {
		fn(&opt)
	}
	return &MockedIdpProvider{ExtSamlMetadata{
		EntityId:         opt.Properties.EntityID,
		Domain:           opt.Properties.Domain,
		Source:           opt.Properties.MetadataSource,
		Name:             opt.Properties.Name,
		IdName:           opt.Properties.IdName,
	}}
}

type MockedIdpProvider struct {
	ExtSamlMetadata
}

func (i MockedIdpProvider) Domain() string {
	return i.ExtSamlMetadata.Domain
}

func (i MockedIdpProvider) GetAutoCreateUserDetails() security.AutoCreateUserDetails {
	return nil
}

func (i MockedIdpProvider) ShouldMetadataRequireSignature() bool {
	return i.ExtSamlMetadata.RequireSignature
}

func (i MockedIdpProvider) ShouldMetadataTrustCheck() bool {
	return i.ExtSamlMetadata.TrustCheck
}

func (i MockedIdpProvider) GetMetadataTrustedKeys() []string {
	return i.ExtSamlMetadata.TrustedKeys
}

func (i MockedIdpProvider) EntityId() string {
	return i.ExtSamlMetadata.EntityId
}

func (i MockedIdpProvider) MetadataLocation() string {
	return i.ExtSamlMetadata.Source
}

func (i MockedIdpProvider) ExternalIdName() string {
	return i.ExtSamlMetadata.IdName
}

func (i MockedIdpProvider) ExternalIdpName() string {
	return i.ExtSamlMetadata.Name
}
