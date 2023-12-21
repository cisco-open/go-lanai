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

package opainput

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/internal"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

func PopulateAuthenticationClause(ctx context.Context, input *opa.Input) error {
	auth := security.Get(ctx)
	if !security.IsFullyAuthenticated(auth) {
		input.Authentication = nil
		return nil
	}
	if input.Authentication == nil {
		input.Authentication = opa.NewAuthenticationClause()
	}
	return populateAuthenticationClause(auth, input.Authentication)
}

func populateAuthenticationClause(auth security.Authentication, clause *opa.AuthenticationClause) error {
	clause.Username = getUsernameOrEmpty(auth)
	clause.Permissions = make([]string, 0, len(auth.Permissions()))
	for k := range auth.Permissions() {
		clause.Permissions = append(clause.Permissions, k)
	}

	switch v := auth.(type) {
	case oauth2.Authentication:
		clause.Client = &opa.OAuthClientClause{
			ClientID:  v.OAuth2Request().ClientId(),
			GrantType: v.OAuth2Request().GrantType(),
			Scopes:    v.OAuth2Request().Scopes().Values(),
		}
	default:
	}

	details := auth.Details()
	if v, ok := details.(security.UserDetails); ok {
		clause.UserID = v.UserId()
	}
	if v, ok := details.(internal.TenantAccessDetails); ok {
		clause.AccessibleTenants = v.EffectiveAssignedTenantIds().Values()
	}
	if v, ok := details.(security.TenantDetails); ok {
		clause.TenantID = v.TenantId()
	}
	if v, ok := details.(security.ProviderDetails); ok {
		clause.ProviderID = v.ProviderId()
	}
	if v, ok := details.(security.AuthenticationDetails); ok {
		clause.Roles = v.Roles().Values()
	}

	return nil
}

func getUsernameOrEmpty(auth security.Authentication) string {
	username, e := security.GetUsername(auth)
	if e != nil {
		return ""
	}
	return username
}
