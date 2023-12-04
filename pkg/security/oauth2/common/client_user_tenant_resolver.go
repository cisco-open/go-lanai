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
