package tenant_hierarchy_modifier

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	tenant_hierarchy_accessor "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenant_hierarchy/accessor"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	r "github.com/go-redis/redis/v8"
)

var logger = log.New("tenant_hierarchy_modifier")

func RemoveTenant(ctx context.Context, tenantId string) error {
	if tenantId == "" {
		return errors.New("tenantId should not be empty")
	}

	logger.Debugf("remove tenantId %s", tenantId)

	children, err := tenant_hierarchy_accessor.GetChildren(ctx, tenantId)

	if err != nil {
		return err
	}

	if len(children) != 0 {
		return errors.New("can't remove tenant that still have children")
	}

	parentId, err := tenant_hierarchy_accessor.GetParent(ctx, tenantId)

	if err != nil {
		return err
	}

	if parentId == nil {
		return errors.New("this tenant is root tenant because it has no parent. root tenant can't be deleted")
	}

	relations := []*r.Z{
		&r.Z{0, tenant_hierarchy_accessor.BuildSpsString(tenantId, tenant_hierarchy_accessor.IsChildrenOfPredict, *parentId)},
		&r.Z{0, tenant_hierarchy_accessor.BuildSpsString(*parentId, tenant_hierarchy_accessor.IsParentOfPredict, tenantId)}}


	cmd := tenant_hierarchy_accessor.RedisClient.ZAdd(ctx, tenant_hierarchy_accessor.ZsetKey, relations...)
	return cmd.Err()
}

func AddTenant(ctx context.Context, tenantId string, parentId string) error {
	if tenantId == "" || parentId == ""{
		return errors.New("tenantId and parentId should not be empty")
	}

	logger.Debugf("add tenantId %s parentId %s", tenantId, parentId)

	ancestors, err := tenant_hierarchy_accessor.GetAnceostors(ctx, parentId)
	if err != nil {
		return err
	}

	set := utils.NewStringSet(ancestors...)

	if set.Has(tenantId) || tenantId == parentId {
		return errors.New("this relationship introduces a cycle in the tenant hierarchy")
	}

	relations := []*r.Z{
		&r.Z{0, tenant_hierarchy_accessor.BuildSpsString(tenantId, tenant_hierarchy_accessor.IsChildrenOfPredict, parentId)},
		&r.Z{0, tenant_hierarchy_accessor.BuildSpsString(parentId, tenant_hierarchy_accessor.IsParentOfPredict, tenantId)}}


	cmd := tenant_hierarchy_accessor.RedisClient.ZAdd(ctx, tenant_hierarchy_accessor.ZsetKey, relations...)
	return cmd.Err()
}
