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

import (
    "encoding/xml"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/crewjam/saml"
)

type MockedClientOptions func(opt *MockedClientOption)
type MockedClientOption struct {
	Properties MockedClientProperties
	SP         *saml.ServiceProvider
}

type MockedSamlClient struct {
	EntityId                             string
	MetadataSource                       string
	SkipAssertionEncryption              bool
	SkipAuthRequestSignatureVerification bool
	MetadataRequireSignature             bool
	MetadataTrustCheck                   bool
	MetadataTrustedKeys                  []string
	TenantRestrictions                   utils.StringSet
	TenantRestrictionType                string
}

func NewMockedSamlClient(opts ...MockedClientOptions) *MockedSamlClient {
	opt := MockedClientOption{}
	for _, fn := range opts {
		fn(&opt)
	}

	if opt.SP != nil {
		metadata := opt.SP.Metadata()
		data, e := xml.Marshal(metadata)
		if e != nil {
			return nil
		}
		return &MockedSamlClient{
			EntityId:              opt.SP.EntityID,
			MetadataSource:        string(data),
			TenantRestrictions:    utils.NewStringSet(),
			TenantRestrictionType: "all",
		}
	}

	return &MockedSamlClient{
		EntityId:                             opt.Properties.EntityID,
		MetadataSource:                       opt.Properties.MetadataSource,
		SkipAssertionEncryption:              opt.Properties.SkipEncryption,
		SkipAuthRequestSignatureVerification: opt.Properties.SkipSignatureVerification,
		TenantRestrictions:                   utils.NewStringSet(opt.Properties.TenantRestriction...),
		TenantRestrictionType:                opt.Properties.TenantRestrictionType,
	}
}

func (c MockedSamlClient) ShouldMetadataRequireSignature() bool {
	return c.MetadataRequireSignature
}

func (c MockedSamlClient) ShouldMetadataTrustCheck() bool {
	return c.MetadataTrustCheck
}

func (c MockedSamlClient) GetMetadataTrustedKeys() []string {
	return c.MetadataTrustedKeys
}

func (c MockedSamlClient) GetEntityId() string {
	return c.EntityId
}

func (c MockedSamlClient) GetMetadataSource() string {
	return c.MetadataSource
}

func (c MockedSamlClient) ShouldSkipAssertionEncryption() bool {
	return c.SkipAssertionEncryption
}

func (c MockedSamlClient) ShouldSkipAuthRequestSignatureVerification() bool {
	return c.SkipAuthRequestSignatureVerification
}

func (c MockedSamlClient) GetTenantRestrictions() utils.StringSet {
	return c.TenantRestrictions
}

func (c MockedSamlClient) GetTenantRestrictionType() string {
	return c.TenantRestrictionType
}
