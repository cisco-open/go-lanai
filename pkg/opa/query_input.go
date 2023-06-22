package opa

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

/********************
	Common Inputs
 ********************/

const (
	InputPrefixRoot           = `input`
	InputPrefixAuthentication = `auth`
	InputPrefixRequest        = `request`
	InputPrefixResource       = `resource`
)

type Input struct {
	Authentication *AuthenticationClause  `json:"auth,omitempty"`
	Request        *RequestClause         `json:"request,omitempty"`
	Resource       *ResourceClause        `json:"resource,omitempty"`
	ExtraData      map[string]interface{} `json:"-"`
}

func (c Input) MarshalJSON() ([]byte, error) {
	type clause Input
	return marshalMergedJSON(clause(c), c.ExtraData)
}

func NewInput() *Input {
	return &Input{}
}

/*****************************
	Common Identity Blocks
 *****************************/

type AuthenticationClause struct {
	// Required fields
	Username    string   `json:"username"`
	Permissions []string `json:"permissions"`
	// Optional fields
	UserID            string                 `json:"user_id,omitempty"`
	TenantID          string                 `json:"tenant_id,omitempty"`
	ProviderID        string                 `json:"provider_id,omitempty"`
	Roles             []string               `json:"roles,omitempty"`
	AccessibleTenants []string               `json:"accessible_tenants,omitempty"`
	Client            *OAuthClientClause     `json:"client"`
	ExtraData         map[string]interface{} `json:"-"`
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

func (c AuthenticationClause) MarshalJSON() ([]byte, error) {
	type clause AuthenticationClause
	return marshalMergedJSON(clause(c), c.ExtraData)
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
	Scheme    string                 `json:"scheme,omitempty"`
	Path      string                 `json:"path,omitempty"`
	Method    string                 `json:"method,omitempty"`
	Header    http.Header            `json:"header,omitempty"`
	Query     url.Values             `json:"query,omitempty"`
	ExtraData map[string]interface{} `json:"-"`
}

func (c RequestClause) MarshalJSON() ([]byte, error) {
	type clause RequestClause
	return marshalMergedJSON(clause(c), c.ExtraData)
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

func (op ResourceOperation) String() string {
	switch op {
	case OpRead:
		return `read`
	case OpWrite:
		return `write`
	case OpCreate:
		return `create`
	case OpDelete:
		return `delete`
	default:
		return ``
	}
}

func (op ResourceOperation) MarshalText() ([]byte, error) {
	return []byte(op.String()), nil
}

func (op ResourceOperation) MarshalJSON() ([]byte, error) {
	text, _ := op.MarshalText()
	return []byte(`"` + string(text) + `"`), nil
}

type ResourceValues struct {
	TenantID   string                 `json:"tenant_id,omitempty"`
	TenantPath []string               `json:"tenant_path,omitempty"`
	OwnerID    string                 `json:"owner_id,omitempty"`
	Share      map[string][]string    `json:"share,omitempty"`
	ExtraData  map[string]interface{} `json:"-"`
}

func (c ResourceValues) MarshalJSON() ([]byte, error) {
	type clause ResourceValues
	return marshalMergedJSON(clause(c), c.ExtraData)
}

type CurrentResourceValues ResourceValues

type ResourceClause struct {
	CurrentResourceValues
	Type      string            `json:"type"`
	Operation ResourceOperation `json:"op"`
	Delta     *ResourceValues    `json:"delta,omitempty"`
}

func NewResourceClause(resType string, op ResourceOperation) *ResourceClause {
	return &ResourceClause{
		Type:      resType,
		Operation: op,
	}
}

func (c ResourceClause) MarshalJSON() ([]byte, error) {
	type clause ResourceClause
	return marshalMergedJSON(clause(c), c.ExtraData)
}

/*************************
	Helpers
 *************************/

// marshalMergedJSON merge extra into v, v have to be struct or map
func marshalMergedJSON(obj interface{}, extra map[string]interface{}) ([]byte, error) {
	data, e := json.Marshal(obj)
	if len(extra) == 0 || e != nil {
		return data, e
	}
	// merge extra
	var m map[string]interface{}
	if e := json.Unmarshal(data, &m); e != nil {
		return nil, fmt.Errorf("unable to merge JSON: %v", e)
	}
	for k, v := range extra {
		m[k] = v
	}
	return json.Marshal(m)
}
