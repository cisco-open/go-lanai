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

const (
	errTmplNotLoaded = `tenancy is not loaded`
)

//goland:noinspection GoNameStartsWithPackageName
type TenancyAccessor struct {
	cachedRootID string
	rc           redis.Client
}

func newAccessor(rc redis.Client) *TenancyAccessor {
	return &TenancyAccessor{
		cachedRootID: "",
		rc:           rc,
	}
}

func (a *TenancyAccessor) GetParent(ctx context.Context, tenantId string) (string, error) {
	if !a.IsLoaded(ctx) {
		return "", errors.New(errTmplNotLoaded)
	}

	gteValue := BuildSpsString(tenantId, IsChildrenOfPredict)
	lteValue := BuildSpsString(tenantId, IsChildrenOfPredict, RedisZsetMaxByte)

	zrange := &r.ZRangeBy{Min: ZInclusive(gteValue), Max: ZInclusive(lteValue)}

	cmd := a.rc.ZRangeByLex(ctx, ZsetKey, zrange)

	relations := cmd.Val()

	if len(relations) == 0 {
		return "", nil
	} else if len(relations) > 1 {
		return "", errors.New(fmt.Sprintf("Tenant should only have one parent, but tenant with Id %s has %d ", tenantId, len(relations)))
	} else {
		p, err := GetObjectOfSpo(relations[0])
		return p, err
	}
}

func (a *TenancyAccessor) GetChildren(ctx context.Context, tenantId string) ([]string, error) {
	if !a.IsLoaded(ctx) {
		return nil, errors.New(errTmplNotLoaded)
	}

	gteValue := BuildSpsString(tenantId, IsParentOfPredict)
	lteValue := BuildSpsString(tenantId, IsParentOfPredict, RedisZsetMaxByte)

	zrange := &r.ZRangeBy{Min: ZInclusive(gteValue), Max: ZInclusive(lteValue)}

	cmd := a.rc.ZRangeByLex(ctx, ZsetKey, zrange)

	var children = make([]string, len(cmd.Val()))
	for i, relation := range cmd.Val() {
		child, err := GetObjectOfSpo(relation)
		if err != nil {
			return nil, err
		}
		children[i] = child
	}
	return children, nil
}

func (a *TenancyAccessor) GetAncestors(ctx context.Context, tenantId string) ([]string, error) {
	if !a.IsLoaded(ctx) {
		return nil, errors.New(errTmplNotLoaded)
	}

	var ancestors = make([]string, 0)
	p, err := a.GetParent(ctx, tenantId)
	for p != "" && err == nil {
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
		return nil, errors.New(errTmplNotLoaded)
	}

	descendants := make([]string, 0)
	idsToVisit := list.New()

	idsToVisit.PushBack(tenantId)

	for idsToVisit.Len() > 0 {

		cmds, err := a.rc.Pipelined(ctx, func(pipeliner r.Pipeliner) error {
			for idsToVisit.Len() > 0 {
				id := idsToVisit.Front()

				gteValue := BuildSpsString(id.Value.(string), IsParentOfPredict)
				lteValue := BuildSpsString(id.Value.(string), IsParentOfPredict, RedisZsetMaxByte)

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

// GetRoot will return the root tenantID.
// Because the root tenantId won't change once system is started, we can cache
// it after first successful read.
func (a *TenancyAccessor) GetRoot(ctx context.Context) (string, error) {
	if !a.IsLoaded(ctx) {
		return "", errors.New(errTmplNotLoaded)
	}
	if a.cachedRootID != "" {
		return a.cachedRootID, nil
	}
	cmd := a.rc.Get(ctx, RootTenantKey)
	if cmd.Err() != nil {
		a.cachedRootID = ""
	} else {
		a.cachedRootID = cmd.Val()
	}
	return a.cachedRootID, cmd.Err()
}

func (a *TenancyAccessor) IsLoaded(ctx context.Context) bool {
	cmd := a.rc.Get(ctx, StatusKey)
	if cmd.Err() != nil {
		return false
	}
	return strings.HasPrefix(cmd.Val(), STATUS_LOADED)
}

func (a *TenancyAccessor) GetTenancyPath(ctx context.Context, tenantId string) ([]uuid.UUID, error) {
	current, err := uuid.Parse(tenantId)
	if err != nil {
		return nil, err
	}
	path := []uuid.UUID{current}

	ancestors, err := a.GetAncestors(ctx, tenantId)
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

	//reverse the order to that the result is root tenant id -> current tenant id
	//fi is index going forward starting from 0,
	//ri is index going backward starting from last element
	//swap the element at ri and ri
	for fi, ri := 0, len(path)-1; fi < ri; fi, ri = fi+1, ri-1 {
		path[fi], path[ri] = path[ri], path[fi]
	}
	return path, nil
}
