package misc

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	httptransport "github.com/go-kit/kit/transport/http"
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

func StringResponseEncoder() httptransport.EncodeResponseFunc {
	return web.CustomResponseEncoder(func(opt *web.EncodeOption) {
		opt.ContentType = "application/json; charset=utf-8"
		opt.WriteFunc = web.TextWriteFunc
	})
}