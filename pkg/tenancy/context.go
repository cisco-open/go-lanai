package tenancy

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/google/uuid"
)

const ZsetKey = "tenant-hierarchy"
const IsParentOfPredict = "is-parent-of"
const IsChildrenOfPredict = "is-children-of"
const RedisZsetMaxByte = "\uffff"
const RootTenantKey = "root-tenant-id"
const StatusKey = "tenant-hierarchy-status"

const STATUS_IN_PROGRESS = "IN_PROGRESS"
const STATUS_LOADED = "LOADED"

var cachedRootId string

type Accessor interface {
	GetParent(ctx context.Context, tenantId string) (string, error)
	GetChildren(ctx context.Context, tenantId string) ([]string, error)
	GetAncestors(ctx context.Context, tenantId string) ([]string, error)
	GetDescendants(ctx context.Context, tenantId string) ([]string, error)
	GetRoot(ctx context.Context) (string, error)
	IsLoaded(ctx context.Context) bool
	GetTenancyPath(ctx context.Context, tenantId string) ([]uuid.UUID, error)
}

// IsLoaded returns if tenancy information is available.
// Note that callers normally don't need to check this flag directly. Other top-level functions Get...() returns error if not loaded
func IsLoaded(ctx context.Context) bool {
	return internalAccessor.IsLoaded(ctx)
}

func GetParent(ctx context.Context, tenantId string) (string, error) {
	return internalAccessor.GetParent(ctx, tenantId)
}

func GetChildren(ctx context.Context, tenantId string) ([]string, error) {
	return internalAccessor.GetChildren(ctx, tenantId)
}

func GetAncestors(ctx context.Context, tenantId string) ([]string, error) {
	return internalAccessor.GetAncestors(ctx, tenantId)
}

func GetDescendants(ctx context.Context, tenantId string) ([]string, error) {
	return internalAccessor.GetDescendants(ctx, tenantId)
}

/*
GetRoot because root tenantId won't change once system is started, we can cache it after first successful read.
 */
func GetRoot(ctx context.Context) (string, error) {
	if cachedRootId != "" {
		return cachedRootId, nil
	}
	var err error
	cachedRootId, err = internalAccessor.GetRoot(ctx)
	return cachedRootId, err
}

func GetTenancyPath(ctx context.Context, tenantId string) ([]uuid.UUID, error) {
	return internalAccessor.GetTenancyPath(ctx, tenantId)
}

// AnyHasDescendant returns true if any of "tenantIDs" in utils.StringSet contains "descendant" or its ancestors
func AnyHasDescendant(ctx context.Context, tenantIDs utils.StringSet, descendant string) bool {
	if tenantIDs == nil || descendant == "" {
		return false
	}

	if tenantIDs.Has(descendant) {
		return true
	}

	ancestors, err := GetAncestors(ctx, descendant)
	if err != nil {
		return false
	}

	for _, ancestor := range ancestors {
		if tenantIDs.Has(ancestor) {
			return true
		}
	}
	return false
}