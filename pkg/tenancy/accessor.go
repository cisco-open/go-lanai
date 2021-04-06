package tenancy

import (
	"container/list"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"errors"
	"fmt"
	r "github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"strings"
)

type TenancyAccessor struct {
	rc redis.Client
}

func newAccessor(rc redis.Client) *TenancyAccessor {
	return &TenancyAccessor{
		rc: rc,
	}
}

func (a *TenancyAccessor) GetParent(ctx context.Context, tenantId string) (string, error) {
	if !a.IsLoaded(ctx) {
		return "", errors.New("tenancy is not loaded")
	}

	gteValue := BuildSpsString(tenantId, IsChildrenOfPredict);
	lteValue := BuildSpsString(tenantId, IsChildrenOfPredict, RedisZsetMaxByte);

	zrange := &r.ZRangeBy{Min: ZInclusive(gteValue), Max: ZInclusive(lteValue)}

	cmd := a.rc.ZRangeByLex(ctx, ZsetKey, zrange)

	relations := cmd.Val()

	if len(relations) == 0 {
		return "", nil;
	} else if len(relations) > 1 {
		return "", errors.New(fmt.Sprintf("Tenant should only have one parent, but tenant with Id %s has %d ", tenantId, len(relations)))
	} else {
		p, err := GetObjectOfSpo(relations[0])
		return p, err
	}
}

func (a *TenancyAccessor) GetChildren(ctx context.Context, tenantId string) ([]string, error) {
	if !a.IsLoaded(ctx) {
		return nil, errors.New("tenancy is not loaded")
	}

	gteValue := BuildSpsString(tenantId, IsParentOfPredict);
	lteValue := BuildSpsString(tenantId, IsParentOfPredict, RedisZsetMaxByte);

	zrange := &r.ZRangeBy{Min: ZInclusive(gteValue), Max: ZInclusive(lteValue)}

	cmd := a.rc.ZRangeByLex(ctx, ZsetKey, zrange)

	var children = []string{}
	for _, relation := range cmd.Val() {
		child, err := GetObjectOfSpo(relation)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return children, nil
}

func (a *TenancyAccessor) GetAnceostors(ctx context.Context, tenantId string) ([]string, error) {
	if !a.IsLoaded(ctx) {
		return nil, errors.New("tenancy is not loaded")
	}

	var ancestors = []string{}

	p, err := a.GetParent(ctx, tenantId)
	for ; p != "" && err == nil; {
		ancestors = append(ancestors, p)
		p, err = a.GetParent(ctx, p)
	}

	if err != nil {
		return nil, err
	}

	return ancestors, nil
}

func (a *TenancyAccessor) GetDescendants(ctx context.Context, tenantId string) ([]string, error) {
	if !a.IsLoaded(ctx) {
		return nil, errors.New("tenancy is not loaded")
	}

	var descendants = []string{}
	idsToVisit := list.New()

	idsToVisit.PushBack(tenantId)

	for idsToVisit.Len()>0 {

		cmds, err := a.rc.Pipelined(ctx, func(pipeliner r.Pipeliner) error {
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

func (a *TenancyAccessor) GetRoot(ctx context.Context) (string, error) {
	if !a.IsLoaded(ctx) {
		return "", errors.New("tenancy is not loaded")
	}

	cmd := a.rc.Get(ctx, RootTenantKey)
	return cmd.Val(), cmd.Err()
}

func (a *TenancyAccessor) IsLoaded(ctx context.Context) bool{
	cmd := a.rc.Get(ctx, StatusKey)
	if cmd.Err() != nil {
		return false
	}
	return	strings.HasPrefix(cmd.Val(), STATUS_LOADED)
}

func (a *TenancyAccessor) GetTenancyPath(ctx context.Context, tenantId string) ([]uuid.UUID, error) {
	current, err := uuid.Parse(tenantId)
	if err != nil {
		return nil, err
	}
	path := []uuid.UUID{current}

	ancestors, err := a.GetAnceostors(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	for _, str := range ancestors {
		id, err := uuid.Parse(str)
		if err != nil {
			return nil, err
		}
		path = append(path, id)
	}
	return path, nil
}