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

package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

// ClientContextDetails implements
// - security.AuthenticationDetails
// - security.KeyValueDetails
// - oauth2.ClientDetails
// It is used to represent a client credential
type ClientContextDetails struct {
	Authentication AuthenticationDetails
	Client         ClientDetails
	KV             map[string]interface{}
	TenantAccess   TenantAccessDetails
}

func (d *ClientContextDetails) ClientId() string {
	return d.Client.Id
}

func (d *ClientContextDetails) AssignedTenantIds() utils.StringSet {
	return d.Client.AssignedTenantIds
}

func (d *ClientContextDetails) Scopes() utils.StringSet {
	return d.Client.Scopes
}

// security.AuthenticationDetails
func (d *ClientContextDetails) ExpiryTime() time.Time {
	return d.Authentication.ExpiryTime
}

// security.AuthenticationDetails
func (d *ClientContextDetails) IssueTime() time.Time {
	return d.Authentication.IssueTime
}

// security.AuthenticationDetails
func (d *ClientContextDetails) Roles() utils.StringSet {
	return d.Authentication.Roles
}

// security.AuthenticationDetails
func (d *ClientContextDetails) Permissions() utils.StringSet {
	return d.Authentication.Permissions
}

// security.AuthenticationDetails
func (d *ClientContextDetails) AuthenticationTime() time.Time {
	return d.Authentication.AuthenticationTime
}

// security.KeyValueDetails
func (d *ClientContextDetails) Value(key string) (v interface{}, ok bool) {
	v, ok = d.KV[key]
	return
}

// security.KeyValueDetails
func (d *ClientContextDetails) Values() (ret map[string]interface{}) {
	ret = map[string]interface{}{}
	for k, v := range d.KV {
		ret[k] = v
	}
	return
}

// ClientTenantedContextDetails implements
// - security.AuthenticationDetails
// - security.KeyValueDetails
// - security.TenantDetails
// - security.ProviderDetails
// - oauth2.ClientDetails
// It is used to represent a client credential with selected tenant
type ClientTenantedContextDetails struct {
	ClientContextDetails
	Tenant   TenantDetails
	Provider ProviderDetails
}

func (d *ClientTenantedContextDetails) TenantId() string {
	return d.Tenant.Id
}

func (d *ClientTenantedContextDetails) TenantExternalId() string {
	return d.Tenant.ExternalId
}

func (d *ClientTenantedContextDetails) TenantSuspended() bool {
	return d.Tenant.Suspended
}

// security.ProviderDetails
func (d *ClientTenantedContextDetails) ProviderId() string {
	return d.Provider.Id
}

// security.ProviderDetails
func (d *ClientTenantedContextDetails) ProviderName() string {
	return d.Provider.Name
}

// security.ProviderDetails
func (d *ClientTenantedContextDetails) ProviderDisplayName() string {
	return d.Provider.DisplayName
}

func (d *ClientTenantedContextDetails) ProviderDescription() string {
	return d.Provider.Description
}

func (d *ClientTenantedContextDetails) ProviderEmail() string {
	return d.Provider.Email
}

func (d *ClientTenantedContextDetails) ProviderNotificationType() string {
	return d.Provider.NotificationType
}
