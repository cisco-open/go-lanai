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

package idp

import (
    "context"
    "errors"
    "github.com/cisco-open/go-lanai/pkg/security"
    "net/http"
    "testing"
)

func TestRequestWithAuthenticationFlow(t *testing.T) {
	idpManager := newTestIdpManager()
	matcher := RequestWithAuthenticationFlow(ExternalIdpSAML, idpManager)

	req, _ := http.NewRequest("GET", "http://192.168.0.1:8900/auth/v2/authorize", nil)
	req.Header.Add("X-Forwarded-Host", "saml.vms.com:443")
	matched, err := matcher.Matches(req)

	if !matched || err != nil {
		t.Errorf("expect to match")
	}
}


type TestIdpProvider struct {
	domain string
	metadataLocation string
	externalIdpName string
	externalIdName string
	entityId string
	metadataRequireSignature bool
	metadataTrustCheck bool
	metadataTrustedKeys []string
}

func (i TestIdpProvider) AuthenticationFlow() AuthenticationFlow {
	return ExternalIdpSAML
}

func (i TestIdpProvider) GetAutoCreateUserDetails() security.AutoCreateUserDetails {
	panic("implement me")
}

func (i TestIdpProvider) ShouldMetadataRequireSignature() bool {
	return i.metadataRequireSignature
}

func (i TestIdpProvider) ShouldMetadataTrustCheck() bool {
	return i.metadataTrustCheck
}

func (i TestIdpProvider) GetMetadataTrustedKeys() []string {
	return i.metadataTrustedKeys
}

func (i TestIdpProvider) Domain() string {
	return i.domain
}

func (i TestIdpProvider) EntityId() string {
	return i.entityId
}

func (i TestIdpProvider) MetadataLocation() string {
	return i.metadataLocation
}

func (i TestIdpProvider) ExternalIdName() string {
	return i.externalIdName
}

func (i TestIdpProvider) ExternalIdpName() string {
	return i.externalIdpName
}

type TestIdpManager struct {
	idpDetails TestIdpProvider
}

func newTestIdpManager() *TestIdpManager {
	return &TestIdpManager{
		idpDetails: TestIdpProvider{
			domain:           "saml.vms.com",
			metadataLocation: "testdata/okta_metadata.xml",
			externalIdpName: "okta",
			externalIdName: "email",
			entityId: "http://www.okta.com/exkwj65c2kC1vwtYi0h7",
		},
	}
}

func (t *TestIdpManager) GetIdentityProvidersWithFlow(context.Context, AuthenticationFlow) []IdentityProvider {
	return []IdentityProvider{t.idpDetails}
}

func (t TestIdpManager) GetIdentityProviderByEntityId(_ context.Context, entityId string) (IdentityProvider, error) {
	if entityId == t.idpDetails.entityId {
		return t.idpDetails, nil
	}
	return nil, errors.New("not found")
}

func (t TestIdpManager) GetIdentityProviderByDomain(_ context.Context, domain string) (IdentityProvider, error) {
	if domain == t.idpDetails.domain {
		return t.idpDetails, nil
	}
	return nil, errors.New("not found")
}