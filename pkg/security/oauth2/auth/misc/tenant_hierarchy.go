package misc

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	tenant_hierarchy_accessor "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenant_hierarchy/accessor"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

type TenantHierarchyEndpoint struct {

}

func NewEndpoint() *TenantHierarchyEndpoint {
	return &TenantHierarchyEndpoint{}
}

type HierarchyRequest struct {
	TenantId string `form:"tenantId"`
}

func (endpoint *TenantHierarchyEndpoint) GetParent(ctx context.Context, req *HierarchyRequest) (string, error) {
	if allow, err := allowAccess(ctx); !allow {
		return "", err
	}

	p, err := tenant_hierarchy_accessor.GetParent(ctx, req.TenantId)
	if err != nil {
		return "", err
	} else if p == nil {
		return "", nil
	} else {
		return *p, err
	}
}

func (endpoint *TenantHierarchyEndpoint) GetChildren(ctx context.Context, req *HierarchyRequest) (utils.StringSet, error) {
	if allow, err := allowAccess(ctx); !allow {
		return nil, err
	}

	children, err := tenant_hierarchy_accessor.GetChildren(ctx, req.TenantId)
	if err == nil {
		ret := utils.NewStringSet(children...)
		return ret, nil
	} else {
		return nil, err
	}
}

func (endpoint *TenantHierarchyEndpoint) GetAncestors(ctx context.Context, req *HierarchyRequest) (utils.StringSet, error) {
	if allow, err := allowAccess(ctx); !allow {
		return nil, err
	}

	ancestor, err := tenant_hierarchy_accessor.GetAnceostors(ctx, req.TenantId)
	if err == nil {
		ret := utils.NewStringSet(ancestor...)
		return ret, nil
	} else {
		return nil, err
	}
}

func (endpoint *TenantHierarchyEndpoint) GetDescendants(ctx context.Context, req *HierarchyRequest) (utils.StringSet, error) {
	if allow, err := allowAccess(ctx); !allow {
		return nil, err
	}

	descendants, err := tenant_hierarchy_accessor.GetDescendants(ctx, req.TenantId)
	if err == nil {
		ret := utils.NewStringSet(descendants...)
		return ret, nil
	} else {
		return nil, err
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