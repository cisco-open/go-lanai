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

package misc

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
	"github.com/cisco-open/go-lanai/pkg/tenancy"
	"github.com/cisco-open/go-lanai/pkg/web"
)

type TenantHierarchyEndpoint struct {

}

func NewTenantHierarchyEndpoint() *TenantHierarchyEndpoint {
	return &TenantHierarchyEndpoint{}
}

type HierarchyRequest struct {
	TenantId string `form:"tenantId"`
}

func (endpoint *TenantHierarchyEndpoint) GetParent(ctx context.Context, req *HierarchyRequest) (string, error) {
	if allow, err := allowAccess(ctx); !allow {
		return "", err
	}

	p, err := tenancy.GetParent(ctx, req.TenantId)
	if err != nil {
		return "", err
	} else {
		return p, err
	}
}

func (endpoint *TenantHierarchyEndpoint) GetChildren(ctx context.Context, req *HierarchyRequest) (interface{}, error) {
	if allow, err := allowAccess(ctx); !allow {
		return nil, err
	}

	children, err := tenancy.GetChildren(ctx, req.TenantId)
	if err == nil {
		ret := children
		return ret, nil
	} else {
		return nil, err
	}
}

func (endpoint *TenantHierarchyEndpoint) GetAncestors(ctx context.Context, req *HierarchyRequest) (interface{}, error) {
	if allow, err := allowAccess(ctx); !allow {
		return nil, err
	}

	ancestor, err := tenancy.GetAncestors(ctx, req.TenantId)
	if err == nil {
		return ancestor, nil
	} else {
		return nil, err
	}
}

func (endpoint *TenantHierarchyEndpoint) GetDescendants(ctx context.Context, req *HierarchyRequest) (interface{}, error) {
	if allow, err := allowAccess(ctx); !allow {
		return nil, err
	}

	descendants, err := tenancy.GetDescendants(ctx, req.TenantId)
	if err == nil {
		ret := descendants
		return ret, nil
	} else {
		return nil, err
	}
}

func (endpoint *TenantHierarchyEndpoint) GetRoot(ctx context.Context, _ *web.EmptyRequest) (string, error) {
	if allow, err := allowAccess(ctx); !allow {
		return "", err
	}

	root, err := tenancy.GetRoot(ctx)

	if err != nil {
		return "", err
	} else {
		return root, nil
	}
}

func allowAccess(ctx context.Context) (bool, error) {
	client := auth.RetrieveAuthenticatedClient(ctx)
	if client == nil {
		return false, oauth2.NewInvalidClientError("tenant hierarchy endpoint requires client authentication")
	}
	if !client.Scopes().Has(oauth2.ScopeTenantHierarchy) {
		return false, oauth2.NewInsufficientScopeError("tenant hierarchy endpoint requires tenant_hierarchy scope")
	}
	return true, nil
}

func StringResponseEncoder() web.EncodeResponseFunc {
	return web.CustomResponseEncoder(func(opt *web.EncodeOption) {
		opt.ContentType = "application/json; charset=utf-8"
		opt.WriteFunc = web.TextWriteFunc
	})
}