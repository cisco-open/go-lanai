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

package samlctx

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/idp"
	"github.com/cisco-open/go-lanai/pkg/utils"
)

/********************
	For IDP
 ********************/

type SamlClient interface {
	GetEntityId() string
	GetMetadataSource() string
	ShouldSkipAssertionEncryption() bool
	ShouldSkipAuthRequestSignatureVerification() bool
	GetTenantRestrictions() utils.StringSet
	GetTenantRestrictionType() string

	ShouldMetadataRequireSignature() bool
	ShouldMetadataTrustCheck() bool
	GetMetadataTrustedKeys() []string
}

type SamlClientStore interface {
	GetAllSamlClient(ctx context.Context) ([]SamlClient, error)
	GetSamlClientByEntityId(ctx context.Context, entityId string) (SamlClient, error)
}

/********************
	For SP
 ********************/

type SamlIdentityProvider interface {
	idp.IdentityProvider
	EntityId() string
	MetadataLocation() string
	ExternalIdName() string
	ExternalIdpName() string
	ShouldMetadataRequireSignature() bool
	ShouldMetadataTrustCheck() bool
	GetMetadataTrustedKeys() []string
	GetAutoCreateUserDetails() security.AutoCreateUserDetails
}

type SamlIdentityProviderManager interface {
	GetIdentityProviderByEntityId(ctx context.Context, entityId string) (idp.IdentityProvider, error)
}

// SamlBindingManager is an additional interface that SamlIdentityProviderManager could implement.
type SamlBindingManager interface {
	// PreferredBindings returns supported bindings in order of preference.
	// possible values are
	// - saml.HTTPRedirectBinding
	// - saml.HTTPPostBinding
	// - saml.HTTPArtifactBinding
	// - saml.SOAPBinding
	// Note that this is not list of supported bindings. Supported bindings are determined by IDP and SP
	PreferredBindings() []string
}