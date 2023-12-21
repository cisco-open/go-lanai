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

package common

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common/internal"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"strings"
	"time"
)

type FactoryOptions func(option *FactoryOption)

type FactoryOption struct {
}

type ContextDetailsFactory struct {
}

func NewContextDetailsFactory(opts ...FactoryOptions) *ContextDetailsFactory {
	opt := FactoryOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &ContextDetailsFactory{}
}

type facts struct {
	request    oauth2.OAuth2Request
	client     oauth2.OAuth2Client
	account    security.Account
	tenant     *security.Tenant
	provider   *security.Provider
	userAuth   oauth2.UserAuthentication
	issueTime  time.Time
	expriyTime time.Time
	authTime   time.Time
	source     oauth2.Authentication
}

func (f *ContextDetailsFactory) New(ctx context.Context, request oauth2.OAuth2Request) (security.ContextDetails, error) {
	facts := f.loadFacts(ctx, request)

	// The auth only have client
	if facts.account == nil {
		return f.createSimple(ctx, facts)
	}

	// The auth has both client and user
	// creates either the ClientUserContextDetail or ClientUserTenantedContextDetail
	return f.create(ctx, facts)
}

/*
*********************

		Helpers
	 *********************
*/
func (f *ContextDetailsFactory) loadFacts(ctx context.Context, request oauth2.OAuth2Request) *facts {
	facts := facts{
		request: request,
		client:  ctx.Value(oauth2.CtxKeyAuthenticatedClient).(oauth2.OAuth2Client),
	}

	if ctx.Value(oauth2.CtxKeyAuthenticatedAccount) != nil {
		facts.account = ctx.Value(oauth2.CtxKeyAuthenticatedAccount).(security.Account)
	}

	if ctx.Value(oauth2.CtxKeyAuthorizedTenant) != nil {
		facts.tenant = ctx.Value(oauth2.CtxKeyAuthorizedTenant).(*security.Tenant)
	}

	if ctx.Value(oauth2.CtxKeyAuthorizedProvider) != nil {
		facts.provider = ctx.Value(oauth2.CtxKeyAuthorizedProvider).(*security.Provider)
	}

	if ctx.Value(oauth2.CtxKeyUserAuthentication) != nil {
		facts.userAuth = ctx.Value(oauth2.CtxKeyUserAuthentication).(oauth2.UserAuthentication)
	}

	if ctx.Value(oauth2.CtxKeyAuthorizationIssueTime) != nil {
		facts.issueTime = ctx.Value(oauth2.CtxKeyAuthorizationIssueTime).(time.Time)
	} else {
		facts.issueTime = time.Now()
	}

	if ctx.Value(oauth2.CtxKeyAuthorizationExpiryTime) != nil {
		facts.expriyTime = ctx.Value(oauth2.CtxKeyAuthorizationExpiryTime).(time.Time)
	}

	if ctx.Value(oauth2.CtxKeyAuthenticationTime) != nil {
		facts.authTime = ctx.Value(oauth2.CtxKeyAuthenticationTime).(time.Time)
	} else {
		facts.authTime = facts.issueTime
	}

	if ctx.Value(oauth2.CtxKeySourceAuthentication) != nil {
		facts.source = ctx.Value(oauth2.CtxKeySourceAuthentication).(oauth2.Authentication)
	}

	return &facts
}

func (f *ContextDetailsFactory) create(ctx context.Context, facts *facts) (security.ContextDetails, error) {
	// user
	ud := internal.UserDetails{
		Id:                facts.account.ID().(string),
		Username:          facts.account.Username(),
		AccountType:       facts.account.Type(),
		AssignedTenantIds: utils.NewStringSet(facts.account.(security.AccountTenancy).DesignatedTenantIds()...),
	}

	if meta, ok := facts.account.(security.AccountMetadata); ok {
		ud.FirstName = meta.FirstName()
		ud.LastName = meta.LastName()
		ud.Email = meta.Email()
		ud.LocaleCode = meta.LocaleCode()
		ud.CurrencyCode = meta.CurrencyCode()
	}

	var cd internal.ClientDetails
	if facts.client != nil {
		cd = internal.ClientDetails{
			Id:                facts.client.ClientId(),
			Scopes:            facts.client.Scopes(),
			AssignedTenantIds: facts.client.AssignedTenantIds(),
		}
	}

	// auth details
	ad, e := f.createAuthDetails(ctx, facts)
	if e != nil {
		return nil, e
	}

	_, assignedTenantId, e := ResolveClientUserTenants(ctx, facts.account, facts.client)
	if e != nil {
		return nil, e
	}

	if facts.tenant != nil {
		// provider
		pd := internal.ProviderDetails{
			Id:               facts.provider.Id,
			Name:             facts.provider.Name,
			DisplayName:      facts.provider.DisplayName,
			Description:      facts.provider.Description,
			Email:            facts.provider.Email,
			NotificationType: facts.provider.NotificationType,
		}

		td := internal.TenantDetails{
			Id:         facts.tenant.Id,
			ExternalId: facts.tenant.ExternalId,
			Suspended:  facts.tenant.Suspended,
		}
		return &internal.ClientUserTenantedContextDetails{
			ClientUserContextDetails: internal.ClientUserContextDetails{
				User:           ud,
				Client:         cd,
				Authentication: *ad,
				KV:             f.createKVDetails(ctx, facts),
				TenantAccess: internal.TenantAccessDetails{
					EffectiveAssignedTenantIds: utils.NewStringSet(assignedTenantId...),
				},
			},
			Provider: pd,
			Tenant:   td,
		}, nil
	} else {
		return &internal.ClientUserContextDetails{
			User:           ud,
			Client:         cd,
			Authentication: *ad,
			KV:             f.createKVDetails(ctx, facts),
			TenantAccess: internal.TenantAccessDetails{
				EffectiveAssignedTenantIds: utils.NewStringSet(assignedTenantId...),
			},
		}, nil
	}
}

func (f *ContextDetailsFactory) createSimple(ctx context.Context, facts *facts) (security.ContextDetails, error) {
	ad, e := f.createAuthDetails(ctx, facts)
	if e != nil {
		return nil, e
	}

	cd := internal.ClientDetails{
		Id:                facts.client.ClientId(),
		Scopes:            facts.client.Scopes(),
		AssignedTenantIds: facts.client.AssignedTenantIds(),
	}

	if facts.tenant != nil {
		td := internal.TenantDetails{
			Id:         facts.tenant.Id,
			ExternalId: facts.tenant.ExternalId,
			Suspended:  facts.tenant.Suspended,
		}
		pd := internal.ProviderDetails{
			Id:               facts.provider.Id,
			Name:             facts.provider.Name,
			DisplayName:      facts.provider.DisplayName,
			Description:      facts.provider.Description,
			NotificationType: facts.provider.NotificationType,
			Email:            facts.provider.Email,
		}
		return &internal.ClientTenantedContextDetails{
			ClientContextDetails: internal.ClientContextDetails{
				Authentication: *ad,
				KV:             f.createKVDetails(ctx, facts),
				Client:         cd,
			},
			Tenant:   td,
			Provider: pd,
		}, nil
	} else {
		return &internal.ClientContextDetails{
			Authentication: *ad,
			KV:             f.createKVDetails(ctx, facts),
			Client:         cd,
		}, nil
	}
}

func (f *ContextDetailsFactory) createAuthDetails(ctx context.Context, facts *facts) (*internal.AuthenticationDetails, error) {
	d := internal.AuthenticationDetails{}
	if facts.account != nil {
		d.Permissions = utils.NewStringSet(facts.account.Permissions()...)
		if meta, ok := facts.account.(security.AccountMetadata); ok {
			d.Roles = utils.NewStringSet(meta.RoleNames()...)
		}
	} else {
		d.Roles = utils.NewStringSet()
		d.Permissions = facts.request.Scopes()
	}

	d.AuthenticationTime = facts.authTime
	d.IssueTime = facts.issueTime
	d.ExpiryTime = facts.expriyTime
	f.populateProxyDetails(ctx, &d, facts)
	return &d, nil
}

func (f *ContextDetailsFactory) populateProxyDetails(_ context.Context, d *internal.AuthenticationDetails, facts *facts) {
	if facts.source == nil {
		return
	}

	if proxyDetails, ok := facts.source.Details().(security.ProxiedUserDetails); ok && proxyDetails.Proxied() {
		// original details is proxied
		d.Proxied = true
		d.OriginalUsername = proxyDetails.OriginalUsername()
		return
	}

	src, ok := facts.source.Details().(security.UserDetails)
	if !ok {
		return
	}

	if facts.account == nil || strings.TrimSpace(facts.account.Username()) != strings.TrimSpace(src.Username()) {
		d.Proxied = true
		d.OriginalUsername = strings.TrimSpace(src.Username())
	}
}

func (f *ContextDetailsFactory) createKVDetails(_ context.Context, facts *facts) (ret map[string]interface{}) {
	ret = map[string]interface{}{}

	if facts.userAuth != nil {
		if sid, ok := facts.userAuth.DetailsMap()[security.DetailsKeySessionId]; ok {
			ret[security.DetailsKeySessionId] = sid
		}
	}

	if facts.request != nil {
		ret[oauth2.DetailsKeyRequestExt] = facts.request.Extensions()
		//ret[oauth2.DetailsKeyRequestParams] = facts.request.Parameters()
	}

	if facts.source == nil {
		return
	}
	if srcKV, ok := facts.source.Details().(security.KeyValueDetails); ok {
		for k, v := range srcKV.Values() {
			ret[k] = v
		}
	}
	return
}
