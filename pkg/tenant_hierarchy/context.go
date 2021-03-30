package tenant_hierarchy

import "context"

type TenantHierarchyStore interface {
	GetIterator(ctx context.Context) TenantIterator
}

type TenantIterator interface {
	Next() bool
	Scan(ctx context.Context) Tenant
	Close() error
	Err() error
}

type Tenant interface {
	GetId() string
	GetParentId() *string //use pointer for nil
}
