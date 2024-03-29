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

package th_loader

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/redis"
    "github.com/cisco-open/go-lanai/pkg/tenancy"
    r "github.com/go-redis/redis/v8"
    "github.com/google/uuid"
)

type TenancyLoader struct {
	rc       redis.Client
	store    TenantHierarchyStore
	accessor tenancy.Accessor
}

func NewLoader(rc redis.Client, store TenantHierarchyStore, accessor tenancy.Accessor) *TenancyLoader {
	return &TenancyLoader{
		rc:       rc,
		store:    store,
		accessor: accessor,
	}
}

func (l *TenancyLoader) LoadTenantHierarchy(ctx context.Context) (err error) {
	//sets status to in progress
	if cmd := l.rc.Set(ctx, tenancy.StatusKey, tenancy.STATUS_IN_PROGRESS, 0); cmd.Err() != nil {
		return cmd.Err()
	}

	//clear out previously loaded data - this way in case the transaction below failed, we get empty data instead of stale data
	if cmd := l.rc.Del(ctx, tenancy.ZsetKey); cmd.Err() != nil {
		return cmd.Err()
	}

	if cmd := l.rc.Del(ctx, tenancy.RootTenantKey); cmd.Err() != nil {
		return cmd.Err()
	}

	//deletes the zset, load its content and set status to loaded_{uuid}
	//this delete is necessary because if two transaction blocks below runs sequentially, we don't want twice the data.
	//this function is to be executed in the transaction below
	var loadTenantHierarchy = func(tx *r.Tx) error {
		cmd := tx.Del(ctx, tenancy.ZsetKey)
		if cmd.Err() != nil {
			return cmd.Err()
		}

		var relations []*r.Z
		it, err := l.store.GetIterator(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = it.Close() }()

		for it.Next() {
			t, err := it.Scan(ctx)

			if err != nil {
				return err
			}

			if t.GetParentId() != "" {
				relations = append(relations,
					&r.Z{Member: tenancy.BuildSpsString(t.GetId(), tenancy.IsChildrenOfPredict, t.GetParentId())},
					&r.Z{Member: tenancy.BuildSpsString(t.GetParentId(), tenancy.IsParentOfPredict, t.GetId())})
			} else {
				statusCmd := tx.Set(ctx, tenancy.RootTenantKey, t.GetId(), 0)
				if statusCmd.Err() != nil {
					return statusCmd.Err()
				}
			}
		}

		// need to wait until root tenant exists
		if cmd := tx.Get(ctx, tenancy.RootTenantKey); cmd.Err() != nil {
			logger.WithContext(ctx).Errorf("Failed to load root tenant due to error: %v", cmd.Err())
			if statusCmd := tx.Set(ctx, tenancy.StatusKey, tenancy.STATUS_FAILED_TO_LOAD_ROOT_TENANT, 0); statusCmd.Err() != nil {
				logger.WithContext(ctx).Errorf("Failed to set status to STATUS_FAILED_TO_LOAD_ROOT_TENANT due to error: %v", statusCmd.Err())
				return statusCmd.Err()
			}
			return cmd.Err()
		}

		if len(relations) != 0 {
			cmd = tx.ZAdd(ctx, tenancy.ZsetKey, relations...)
			if cmd.Err() != nil {
				return cmd.Err()
			}
		}

		statusCmd := tx.Set(ctx, tenancy.StatusKey, fmt.Sprintf("%s_%s", tenancy.STATUS_LOADED, uuid.New().String()), 0)
		if statusCmd.Err() != nil {
			return statusCmd.Err()
		}
		return nil
	}

	//watches the status key, if status changed, the transaction is aborted
	err = l.rc.Watch(ctx, loadTenantHierarchy, tenancy.StatusKey)

	if err != nil {
		//we check if the failure is due to transaction aborted.
		//if the status is loaded, that means this process failed because another auth server instance has loaded the
		//successfully at the same time. If that's the case, we can start up.
		loaded := l.accessor.IsLoaded(ctx)
		if loaded {
			return nil
		}
	}
	return
}
