package th_loader

import "context"

type TenantHierarchyStore interface {
	GetIterator(ctx context.Context) (TenantIterator, error)
}

type TenantIterator interface {
	Next() bool
	Scan(ctx context.Context) (Tenant, error)
	Close() error
	Err() error
}

type Tenant interface {
	GetId() string
	GetParentId() *string //use pointer for nil
}

type Loader interface {
	LoadTenantHierarchy(ctx context.Context) (err error)
}

func LoadTenantHierarchy(ctx context.Context) (err error) {
	return internalLoader.LoadTenantHierarchy(ctx)
}