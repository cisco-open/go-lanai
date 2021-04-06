package th_modifier

import (
	"context"
)

type Modifier interface {
	RemoveTenant(ctx context.Context, tenantId string) error
	AddTenant(ctx context.Context, tenantId string, parentId string) error
}

func RemoveTenant(ctx context.Context, tenantId string) error {
	return internaModifier.RemoveTenant(ctx, tenantId)
}
func AddTenant(ctx context.Context, tenantId string, parentId string) error {
	return internaModifier.AddTenant(ctx, tenantId, parentId)
}
