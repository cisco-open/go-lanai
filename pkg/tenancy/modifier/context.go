package th_modifier

import (
	"context"
)

type Modifier interface {
	RemoveTenant(ctx context.Context, tenantId string) error
	AddTenant(ctx context.Context, tenantId string, parentId string) error
}

func RemoveTenant(ctx context.Context, tenantId string) error {
	return internalModifier.RemoveTenant(ctx, tenantId)
}
func AddTenant(ctx context.Context, tenantId string, parentId string) error {
	return internalModifier.AddTenant(ctx, tenantId, parentId)
}
