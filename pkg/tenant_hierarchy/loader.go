package tenant_hierarchy

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	tenant_hierarchy_accessor "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenant_hierarchy/accessor"
	"fmt"
	r "github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

func LoadTenantHierarchy(ctx context.Context, store TenantHierarchyStore, redisClient redis.Client) (err error) {
	//sets status to in progress
	if cmd := redisClient.Set(ctx, tenant_hierarchy_accessor.StatusKey, tenant_hierarchy_accessor.STATUS_IN_PROGRESS, 0); cmd.Err() != nil {
		err = cmd.Err()
		return
	}

	//clear out previously loaded data - this way in case the transaction below failed, we get empty data instead of stale data
	if cmd := redisClient.Del(ctx, tenant_hierarchy_accessor.ZsetKey); cmd.Err() != nil {
		err = cmd.Err()
		return
	}

	if cmd := redisClient.Del(ctx, tenant_hierarchy_accessor.RootTenantKey); cmd.Err() != nil {
		err = cmd.Err()
		return
	}

	//deletes the zset, load its content and set status to loaded_{uuid}
	//this delete is necessary because if two transaction blocks below runs sequentially, we don't want twice the data.
	//this function is to be executed in the transaction below
	var loadTenantHierarchy = func(tx *r.Tx) error {
		cmd := tx.Del(ctx, tenant_hierarchy_accessor.ZsetKey)
		if cmd.Err() != nil {
			return cmd.Err()
		}

		var relations []*r.Z
		for it := store.GetIterator(ctx); it.Next(); {
			t := it.Scan(ctx)

			if t.GetParentId() != nil {
				relations = append(relations,
					&r.Z{0, tenant_hierarchy_accessor.BuildSpsString(t.GetId(), tenant_hierarchy_accessor.IsChildrenOfPredict, *t.GetParentId())},
					&r.Z{0, tenant_hierarchy_accessor.BuildSpsString(*t.GetParentId(), tenant_hierarchy_accessor.IsParentOfPredict, t.GetId())})
			}

			if t.GetParentId() == nil {
				statusCmd := tx.Set(ctx, tenant_hierarchy_accessor.RootTenantKey, t.GetId(), 0)
				if statusCmd.Err() != nil {
					return statusCmd.Err()
				}
			}
		}

		cmd = tx.ZAdd(ctx, tenant_hierarchy_accessor.ZsetKey, relations...)
		if cmd.Err() != nil {
			return cmd.Err()
		}

		statusCmd := tx.Set(ctx, tenant_hierarchy_accessor.StatusKey, fmt.Sprintf("%s_%s", tenant_hierarchy_accessor.STATUS_LOADED, uuid.New().String()), 0)
		if statusCmd.Err() != nil {
			return statusCmd.Err()
		}
		return nil
	}

	//watches the status key, if status changed, the transaction is aborted
	err = redisClient.Watch(ctx, loadTenantHierarchy, tenant_hierarchy_accessor.StatusKey)
	return
}