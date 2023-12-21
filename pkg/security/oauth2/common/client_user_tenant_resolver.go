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

package common

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"golang.org/x/exp/slices"
)

/*
ResolveClientUserTenants will take the client's assigned tenants and the user's assigned tenants, and use them to compute the tenants
this security context has access to as a result. For example, if a client is assigned to tenant-1, it means anyone using this client
has access to tenant-1. If a user who has access to tenant-1 and tenant-2 is authenticated using this client. Then the resulting
security context should indicate that the user has only access to tenant-1. As a result, the user's default tenant may or may not
still be valid, so this method also returns that.
*/
func ResolveClientUserTenants(ctx context.Context, a security.Account, c oauth2.OAuth2Client) (defaultTenantId string, assignedTenants []string, err error) {
	// client only
	if a == nil {
		assignedTenants = c.AssignedTenantIds().Values()
		if len(assignedTenants) == 1 {
			defaultTenantId = assignedTenants[0]
		}
		return defaultTenantId, assignedTenants, nil
	}

	at, ok := a.(security.AccountTenancy)
	if !ok {
		return "", nil, errors.New("account must have tenancy")
	}

	// To get the intersection of client and user's tenants
	// we need to do two loops.
	// First loop through the account's tenant.
	// If this tenant is any of the client's tenant's descendant, we add it to the return set.
	// Then loop through the client's tenant.
	// If this tenant is any of the account's tenant's descendant, we add it to the return set.
	tenantSet := utils.NewStringSet()
	if c.AssignedTenantIds().Has(security.SpecialTenantIdWildcard) {
		tenantSet = tenantSet.Add(at.DesignatedTenantIds()...)
	} else {
		for _, t := range at.DesignatedTenantIds() {
			if tenancy.AnyHasDescendant(ctx, c.AssignedTenantIds(), t) {
				tenantSet = tenantSet.Add(t)
			}
		}
	}

	if slices.Contains(at.DesignatedTenantIds(), security.SpecialTenantIdWildcard) {
		tenantSet = tenantSet.Add(c.AssignedTenantIds().Values()...)
	} else {
		for t, _ := range c.AssignedTenantIds() {
			if tenancy.AnyHasDescendant(ctx,
				utils.NewStringSet(at.DesignatedTenantIds()...), t) {
				tenantSet = tenantSet.Add(t)
			}
		}
	}

	if tenantSet.Has(security.SpecialTenantIdWildcard) ||
		tenancy.AnyHasDescendant(ctx, tenantSet, a.(security.AccountTenancy).DefaultDesignatedTenantId()) {
		defaultTenantId = a.(security.AccountTenancy).DefaultDesignatedTenantId()
	}

	assignedTenants = tenantSet.Values()
	return defaultTenantId, assignedTenants, nil
}
