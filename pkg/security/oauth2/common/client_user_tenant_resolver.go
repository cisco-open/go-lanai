package common

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
)

func ResolveClientUserTenants(ctx context.Context, a security.Account, c oauth2.OAuth2Client) (defaultTenantId string, assignedTenants []string, err error) {
	if _, ok := a.(security.AccountTenancy); !ok {
		return "", nil, errors.New("account must have tenancy")
	}

	for _, t := range a.(security.AccountTenancy).DesignatedTenantIds() {
		if c.Scopes().Has(oauth2.ScopeCrossTenant) || tenancy.AnyHasDescendant(ctx, c.AssignedTenantIds(), t) {
			assignedTenants = append(assignedTenants, t)
		}
	}

	if tenancy.AnyHasDescendant(ctx,
		utils.NewStringSet(assignedTenants...),
		a.(security.AccountTenancy).DefaultDesignatedTenantId()) {
		defaultTenantId = a.(security.AccountTenancy).DefaultDesignatedTenantId()
	}

	return defaultTenantId, assignedTenants, nil
}
