package th_modifier

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	r "github.com/go-redis/redis/v8"
)

type TenancyModifer struct {
	rc       redis.Client
	accessor tenancy.Accessor
}

func newModifier(rc redis.Client, accessor tenancy.Accessor) *TenancyModifer {
	return &TenancyModifer{
		rc:       rc,
		accessor: accessor,
	}
}

func (m *TenancyModifer) RemoveTenant(ctx context.Context, tenantId string) error {
	if tenantId == "" {
		return errors.New("tenantId should not be empty")
	}

	logger.Debugf("remove tenantId %s", tenantId)

	children, err := m.accessor.GetChildren(ctx, tenantId)

	if err != nil {
		return err
	}

	if len(children) != 0 {
		return errors.New("can't remove tenant that still have children")
	}

	parentId, err := m.accessor.GetParent(ctx, tenantId)

	if err != nil {
		return err
	}

	if parentId == "" {
		return errors.New("this tenant is root tenant because it has no parent. root tenant can't be deleted")
	}

	relations := []interface{}{
		tenancy.BuildSpsString(tenantId, tenancy.IsChildrenOfPredict, parentId),
		tenancy.BuildSpsString(parentId, tenancy.IsParentOfPredict, tenantId)}

	cmd := m.rc.ZRem(ctx, tenancy.ZsetKey, relations...)
	return cmd.Err()
}

func (m *TenancyModifer) AddTenant(ctx context.Context, tenantId string, parentId string) error {
	if tenantId == "" || parentId == "" {
		return errors.New("tenantId and parentId should not be empty")
	}

	logger.Debugf("add tenantId %s parentId %s", tenantId, parentId)

	p, err := m.accessor.GetParent(ctx, tenantId)
	if err != nil {
		return err
	}
	if p != "" {
		return errors.New("this tenant already have a parent")
	}

	root, err := m.accessor.GetRoot(ctx)
	if err != nil {
		return err
	}
	if tenantId == root {
		return errors.New("this tenant is the root")
	}

	ancestors, err := m.accessor.GetAncestors(ctx, parentId)
	if err != nil {
		return err
	}

	set := utils.NewStringSet(ancestors...)

	if set.Has(tenantId) || tenantId == parentId {
		return errors.New("this relationship introduces a cycle in the tenant hierarchy")
	}

	relations := []*r.Z{
		{Member: tenancy.BuildSpsString(tenantId, tenancy.IsChildrenOfPredict, parentId)},
		{Member: tenancy.BuildSpsString(parentId, tenancy.IsParentOfPredict, tenantId)}}

	cmd := m.rc.ZAdd(ctx, tenancy.ZsetKey, relations...)
	return cmd.Err()
}
