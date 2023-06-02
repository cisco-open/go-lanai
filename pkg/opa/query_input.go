package opa

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"net/http"
	"net/url"
)

/********************
	Common Inputs
 ********************/

type InputApiAccess struct {
	Authentication *AuthenticationClause `json:"auth,omitempty"`
	Request        *RequestClause        `json:"request,omitempty"`
	Resource       *ResourceClause       `json:"resource,omitempty"`
}

/*****************************
	Common Identity Blocks
 *****************************/

type AuthenticationClause struct {
	// Required fields
	Username    string   `json:"username"`
	Permissions []string `json:"permissions"`
	// Optional fields
	UserID            string             `json:"user_id,omitempty"`
	TenantID          string             `json:"tenant_id,omitempty"`
	ProviderID        string             `json:"provider_id,omitempty"`
	Roles             []string           `json:"roles,omitempty"`
	AccessibleTenants []string           `json:"accessible_tenants,omitempty"`
	Client            *OAuthClientClause `json:"client"`
}

type OAuthClientClause struct {
	ClientID  string   `json:"client_id"`
	GrantType string   `json:"grant_type,omitempty"`
	Scopes    []string `json:"scopes"`
}

func NewAuthenticationClause(ctx context.Context) *AuthenticationClause {
	auth := security.Get(ctx)
	ret := newAuthenticationClause(auth)
	details := auth.Details()
	if v, ok := details.(security.UserDetails); ok {
		ret.UserID = v.UserId()
		ret.AccessibleTenants = v.AssignedTenantIds().Values()
	}
	if v, ok := details.(security.TenantDetails); ok {
		ret.TenantID = v.TenantId()
	}
	if v, ok := details.(security.ProviderDetails); ok {
		ret.ProviderID = v.ProviderId()
	}
	if v, ok := details.(security.AuthenticationDetails); ok {
		ret.Roles = v.Roles().Values()
	}

	return ret
}

func newAuthenticationClause(auth security.Authentication) *AuthenticationClause {
	ret := AuthenticationClause{
		Username: mustGetUsername(auth),
	}
	ret.Permissions = make([]string, 0, len(auth.Permissions()))
	for k := range auth.Permissions() {
		ret.Permissions = append(ret.Permissions, k)
	}

	switch v := auth.(type) {
	case oauth2.Authentication:
		ret.Client = &OAuthClientClause{
			ClientID:  v.OAuth2Request().ClientId(),
			GrantType: v.OAuth2Request().GrantType(),
			Scopes:    v.OAuth2Request().Scopes().Values(),
		}
	default:
	}
	return &ret
}

func mustGetUsername(auth security.Authentication) string {
	username, e := security.GetUsername(auth)
	if e != nil {
		return ""
	}
	return username
}

/**************************
	Common Resource Blocks
 **************************/

type RequestClause struct {
	Scheme string      `json:"scheme,omitempty"`
	Path   string      `json:"path,omitempty"`
	Method string      `json:"method,omitempty"`
	Header http.Header `json:"header,omitempty"`
	Query  url.Values  `json:"query,omitempty"`
}

func NewRequestClause(req *http.Request) *RequestClause {
	return &RequestClause{
		Scheme: req.URL.Scheme,
		Path:   req.URL.Path,
		Method: req.Method,
		Header: req.Header,
		Query:  req.URL.Query(),
	}
}

type ResourceOperation int

const (
	OpRead ResourceOperation = iota
	OpWrite
	OpCreate
	OpDelete
)

func (op ResourceOperation) MarshalText() ([]byte, error) {
	switch op {
	case OpRead:
		return []byte(`read`), nil
	case OpWrite:
		return []byte(`write`), nil
	case OpCreate:
		return []byte("create"), nil
	case OpDelete:
		return []byte("delete"), nil
	default:
		return []byte{}, nil
	}
}

func (op ResourceOperation) MarshalJSON() ([]byte, error) {
	text, _ := op.MarshalText()
	return []byte(`"` + string(text) + `"`), nil
}

type ResourceClause struct {
	Type       string              `json:"type"`
	Operation  ResourceOperation   `json:"op"`
	TenantID   string              `json:"tenant_id,omitempty"`
	TenantPath []string            `json:"tenant_path,omitempty"`
	OwnerID    string              `json:"owner_id,omitempty"`
	Share      map[string][]string `json:"share,omitempty"`
}

func NewResourceClause(resType string, op ResourceOperation) *ResourceClause {
	return &ResourceClause{
		Type: resType,
		Operation: op,
	}
}
