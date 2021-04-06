package tenancy

import "context"

const ZsetKey = "tenant-hierarchy"
const IsParentOfPredict = "is-parent-of"
const IsChildrenOfPredict = "is-children-of"
const RedisZsetMaxByte = "\uffff"
const RootTenantKey = "root-tenant-id"
const StatusKey = "tenant-hierarchy-status"

const STATUS_IN_PROGRESS = "IN_PROGRESS"
const STATUS_LOADED = "LOADED"

type Accessor interface {
	GetParent(ctx context.Context, tenantId string) (string, error)
	GetChildren(ctx context.Context, tenantId string) ([]string, error)
	GetAnceostors(ctx context.Context, tenantId string) ([]string, error)
	GetDescendants(ctx context.Context, tenantId string) ([]string, error)
	GetRoot(ctx context.Context) (string, error)
	IsLoaded(ctx context.Context) bool
}

func GetParent(ctx context.Context, tenantId string) (string, error) {
	return internalAccessor.GetParent(ctx, tenantId)
}
func GetChildren(ctx context.Context, tenantId string) ([]string, error) {
	return internalAccessor.GetChildren(ctx, tenantId)
}
func GetAnceostors(ctx context.Context, tenantId string) ([]string, error) {
	return internalAccessor.GetAnceostors(ctx, tenantId)
}
func GetDescendants(ctx context.Context, tenantId string) ([]string, error) {
	return internalAccessor.GetDescendants(ctx, tenantId)
}
func GetRoot(ctx context.Context) (string, error) {
	return internalAccessor.GetRoot(ctx)
}
func IsLoaded(ctx context.Context) bool {
	return internalAccessor.IsLoaded(ctx)
}