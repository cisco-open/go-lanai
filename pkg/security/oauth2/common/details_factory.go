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

func NewContextDetailsFactory(opts...FactoryOptions) *ContextDetailsFactory {
	opt := FactoryOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &ContextDetailsFactory {

	}
}

type facts struct {
	request    oauth2.OAuth2Request
	client     oauth2.OAuth2Client
	account    security.Account
	tenant     *security.Tenant
	provider   *security.Provider
	issueTime  time.Time
	expriyTime time.Time
	authTime   time.Time
	source     oauth2.Authentication
}

func (f *ContextDetailsFactory) New(ctx context.Context, request oauth2.OAuth2Request) (security.ContextDetails, error) {
	facts := f.loadFacts(ctx, request)
	if facts.account == nil || facts.tenant == nil || facts.provider == nil {
		return f.createSimple(ctx, facts)
	}
	return f.create(ctx, facts)
}

/**********************
	Helpers
 **********************/
func (f *ContextDetailsFactory) loadFacts(ctx context.Context, request oauth2.OAuth2Request) *facts {
	facts := facts{
		request: request,
		client: ctx.Value(oauth2.CtxKeyAuthenticatedClient).(oauth2.OAuth2Client),
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

func (f *ContextDetailsFactory) create(ctx context.Context, facts *facts) (*internal.FullContextDetails, error) {
	// provider
	pd := internal.ProviderDetails{
		Id: facts.provider.Id,
		Name: facts.provider.Name,
		DisplayName: facts.provider.DisplayName,
	}

	// tenant
	td := internal.TenantDetails{
		Id: facts.tenant.Id,
		Name: facts.tenant.Name,
		Suspended: facts.tenant.Suspended,
	}

	// user
	ud := internal.UserDetails{
		Id: facts.account.ID().(string),
		Username: facts.account.Username(),
		AccountType: facts.account.Type(),
		AssignedTenantIds: utils.NewStringSet(facts.account.(security.AccountTenancy).Tenants()...),
	}

	if meta, ok := facts.account.(security.AccountMetadata); ok {
		ud.FirstName = meta.FirstName()
		ud.LastName = meta.LastName()
		ud.Email = meta.Email()
		ud.LocaleCode = meta.LocaleCode()
		ud.CurrencyCode = meta.CurrencyCode()
	}

	// creds
	ad, e := f.createAuthDetails(ctx, facts)
	if e != nil {
		return nil, e
	}

	return &internal.FullContextDetails{
		Provider: pd,
		Tenant: td,
		User: ud,
		Authentication: *ad,
		KV: map[string]interface{}{},
	},  nil
}

func (f *ContextDetailsFactory) createSimple(ctx context.Context, facts *facts) (*internal.SimpleContextDetails, error) {
	// creds
	ad, e := f.createAuthDetails(ctx, facts) // TODO
	if e != nil {
		return nil, e
	}

	return &internal.SimpleContextDetails{
		Authentication: *ad,
		KV: map[string]interface{}{},
	},  nil
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
		d.Permissions = facts.client.Scopes().Copy()
	}

	d.AuthenticationTime = facts.authTime
	d.IssueTime = facts.issueTime
	d.ExpiryTime = facts.expriyTime
	f.populateProxyDetails(ctx, &d, facts)
	return &d, nil
}

func (f *ContextDetailsFactory) populateProxyDetails(ctx context.Context, d *internal.AuthenticationDetails, facts *facts) {
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