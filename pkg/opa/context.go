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

package opa

import (
	"context"
	"net/http"
	"net/url"
)

/********************
	Constants
 ********************/

const (
	PackagePrefixResource = `resource`
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
	return &Input{
		ExtraData: make(map[string]interface{}),
	}
}

type InputCustomizer interface {
	Customize(ctx context.Context, input *Input) error
}

type InputCustomizerFunc func(ctx context.Context, input *Input) error

func (fn InputCustomizerFunc) Customize(ctx context.Context, input *Input) error {
	return fn(ctx, input)
}

/*****************************
	Common Identity Blocks
 *****************************/

type AuthenticationClause struct {
	// Required fields
	UserID      string   `json:"user_id"`
	Permissions []string `json:"permissions"`
	// Optional fields
	Username          string                 `json:"username,omitempty"`
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

func (c AuthenticationClause) MarshalJSON() ([]byte, error) {
	type clause AuthenticationClause
	return marshalMergedJSON(clause(c), c.ExtraData)
}

func NewAuthenticationClause() *AuthenticationClause {
	return &AuthenticationClause{
		ExtraData: map[string]interface{}{},
	}
}

/**************************
	Common ResourceQuery Blocks
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

type ResourceOperation string

const (
	OpRead   ResourceOperation = `read`
	OpWrite  ResourceOperation = `write`
	OpCreate ResourceOperation = `create`
	OpDelete ResourceOperation = `delete`
)

type ResourceValues struct {
	TenantID   string                         `json:"tenant_id,omitempty"`
	TenantPath []string                       `json:"tenant_path,omitempty"`
	OwnerID    string                         `json:"owner_id,omitempty"`
	Sharing    map[string][]ResourceOperation `json:"sharing,omitempty"`
	ExtraData  map[string]interface{}         `json:"-"`
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
	Delta     *ResourceValues   `json:"delta,omitempty"`
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
