package tenant_hierarchy_accessor

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	r "github.com/go-redis/redis/v8"
)

const ZsetKey = "tenant-hierarchy"
const IsParentOfPredict = "is-parent-of"
const IsChildrenOfPredict = "is-children-of"
const RedisZsetMaxByte = "\uffff"
const RootTenantKey = "root-tenant-id"
const StatusKey = "tenant-hierarchy-status"

const STATUS_IN_PROGRESS = "IN_PROGRESS"
const STATUS_LOADED = "LOADED"
const STATUS_ERROR = "ERROR"

func GetParent(ctx context.Context, tenantId string) (*string, error) {
	gteValue := BuildSpsString(tenantId, IsChildrenOfPredict);
	lteValue := BuildSpsString(tenantId, IsChildrenOfPredict, RedisZsetMaxByte);

	zrange := &r.ZRangeBy{Min: ZInclusive(gteValue), Max: ZInclusive(lteValue)}

	cmd := RedisClient.ZRangeByLex(ctx, ZsetKey, zrange)

	relations := cmd.Val()

	if len(relations) == 0 {
		return nil, nil;
	} else if len(relations) > 1 {
		return nil, errors.New(fmt.Sprintf("Tenant should only have one parent, but tenant with Id %s has %d ", tenantId, len(relations)))
	} else {
		p, err := GetObjectOfSpo(relations[0])
		return &p, err
	}
}

func GetChildren(ctx context.Context, tenantId string) ([]string, error) {
	gteValue := BuildSpsString(tenantId, IsParentOfPredict);
	lteValue := BuildSpsString(tenantId, IsParentOfPredict, RedisZsetMaxByte);

	zrange := &r.ZRangeBy{Min: ZInclusive(gteValue), Max: ZInclusive(lteValue)}

	cmd := RedisClient.ZRangeByLex(ctx, ZsetKey, zrange)

	var children []string
	for _, relation := range cmd.Val() {
		child, err := GetObjectOfSpo(relation)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return children, nil
}

func GetAnceostors(ctx context.Context, tenantId string) ([]string, error) {
	var ancestors []string

	p, err := GetParent(ctx, tenantId)
	for ; p != nil && err == nil; {
		ancestors = append(ancestors, *p)
		p, err = GetParent(ctx, *p)
	}

	if err != nil {
		return nil, err
	}

	return ancestors, nil
}

func GetDescendants(ctx context.Context, tenantId string) ([]string, error) {
	var descendants []string
	idsToVisit := list.New()

	idsToVisit.PushBack(tenantId)

	for idsToVisit.Len()>0 {

		cmds, err := RedisClient.Pipelined(ctx, func(pipeliner r.Pipeliner) error {
			for idsToVisit.Len() > 0 {
				id := idsToVisit.Front()

				gteValue := BuildSpsString(id.Value.(string), IsParentOfPredict);
				lteValue := BuildSpsString(id.Value.(string), IsParentOfPredict, RedisZsetMaxByte);

				zrange := &r.ZRangeBy{Min: ZInclusive(gteValue), Max: ZInclusive(lteValue)}

				pcmd := pipeliner.ZRangeByLex(ctx, ZsetKey, zrange)

				if pcmd.Err() == nil {
					idsToVisit.Remove(id)
				} else {
					return pcmd.Err()
				}
			}
			return nil
		})

		if err != nil {
			return nil, err
		}

		var children []string
		for _, c := range cmds {
			for _, relation := range c.(*r.StringSliceCmd).Val() {
				child, err := GetObjectOfSpo(relation)
				if err != nil {
					return nil, err
				}
				children = append(children, child)
			}
		}

		descendants = append(descendants, children...)
		for _, child := range children {
			idsToVisit.PushBack(child)
		}
	}
	return descendants, nil
}

func GetRoot(ctx context.Context) (string, error) {
	cmd := RedisClient.Get(ctx, RootTenantKey)
	return cmd.Val(), cmd.Err()
}

func IsLoaded(ctx context.Context) bool {
	cmd := RedisClient.Get(ctx, StatusKey)
	return cmd.Val() == STATUS_LOADED
}