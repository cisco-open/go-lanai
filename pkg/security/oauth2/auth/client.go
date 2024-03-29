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

package auth

import (
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"time"
)

/***********************************
	Default implementation
 ***********************************/

type ClientDetails struct {
	ClientId             string
	Secret               string
	GrantTypes           utils.StringSet
	RedirectUris         utils.StringSet
	Scopes               utils.StringSet
	AutoApproveScopes    utils.StringSet
	AccessTokenValidity  time.Duration
	RefreshTokenValidity time.Duration
	UseSessionTimeout    bool
	AssignedTenantIds    utils.StringSet
	ResourceIds          utils.StringSet
}

// DefaultOAuth2Client implements security.Account & OAuth2Client
type DefaultOAuth2Client struct {
	ClientDetails
}

// deja vu
func NewClient() *DefaultOAuth2Client {
	return &DefaultOAuth2Client{}
}

func NewClientWithDetails(clientDetails ClientDetails) *DefaultOAuth2Client {
	return &DefaultOAuth2Client{
		ClientDetails: clientDetails,
	}
}

/** OAuth2Client **/
func (c *DefaultOAuth2Client) ClientId() string {
	return c.ClientDetails.ClientId
}

func (c *DefaultOAuth2Client) SecretRequired() bool {
	return c.ClientDetails.Secret != ""
}

func (c *DefaultOAuth2Client) Secret() string {
	return c.ClientDetails.Secret
}

func (c *DefaultOAuth2Client) GrantTypes() utils.StringSet {
	return c.ClientDetails.GrantTypes
}

func (c *DefaultOAuth2Client) RedirectUris() utils.StringSet {
	return c.ClientDetails.RedirectUris
}

func (c *DefaultOAuth2Client) Scopes() utils.StringSet {
	return c.ClientDetails.Scopes
}

func (c *DefaultOAuth2Client) AutoApproveScopes() utils.StringSet {
	return c.ClientDetails.AutoApproveScopes
}

func (c *DefaultOAuth2Client) AccessTokenValidity() time.Duration {
	return c.ClientDetails.AccessTokenValidity
}

func (c *DefaultOAuth2Client) RefreshTokenValidity() time.Duration {
	return c.ClientDetails.RefreshTokenValidity
}

func (c *DefaultOAuth2Client) UseSessionTimeout() bool {
	return c.ClientDetails.UseSessionTimeout
}

func (c *DefaultOAuth2Client) AssignedTenantIds() utils.StringSet {
	return c.ClientDetails.AssignedTenantIds
}

func (c *DefaultOAuth2Client) ResourceIDs() utils.StringSet {
	return c.ClientDetails.ResourceIds
}

func (c *DefaultOAuth2Client) MaxTokensPerUser() int {
	return -1
}

/** security.Account **/

func (c *DefaultOAuth2Client) ID() interface{} {
	return c.ClientDetails.ClientId
}

func (c *DefaultOAuth2Client) Type() security.AccountType {
	return security.AccountTypeDefault
}

func (c *DefaultOAuth2Client) Username() string {
	return c.ClientDetails.ClientId
}

func (c *DefaultOAuth2Client) Credentials() interface{} {
	return c.ClientDetails.Secret
}

func (c *DefaultOAuth2Client) Permissions() []string {
	return c.ClientDetails.Scopes.Values()
}

func (c *DefaultOAuth2Client) Disabled() bool {
	return false
}

func (c *DefaultOAuth2Client) Locked() bool {
	return false
}

func (c *DefaultOAuth2Client) UseMFA() bool {
	return false
}

func (c *DefaultOAuth2Client) CacheableCopy() security.Account {
	copy := DefaultOAuth2Client{
		ClientDetails: c.ClientDetails,
	}
	copy.ClientDetails.Secret = ""
	return &copy
}
