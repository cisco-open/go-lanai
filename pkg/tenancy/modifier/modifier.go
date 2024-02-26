// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package th_modifier

import (
    "context"
    "errors"
    "github.com/cisco-open/go-lanai/pkg/redis"
    "github.com/cisco-open/go-lanai/pkg/tenancy"
    "github.com/cisco-open/go-lanai/pkg/utils"
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
