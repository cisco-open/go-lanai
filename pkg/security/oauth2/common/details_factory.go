package common

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common/internal"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"time"
)

type FactoryOptions func(option *FactoryOption)

type FactoryOption struct {
	ClientStore    oauth2.OAuth2ClientStore
	AccountStore   security.AccountStore
	TenantStore    security.TenantStore
	ProviderStore  security.ProviderStore
	HierarchyStore security.TenantHierarchyStore
}

type ContextDetailsFactory struct {
	clientStore    oauth2.OAuth2ClientStore
	accountStore   security.AccountStore
	tenantStore    security.TenantStore
	providerStore  security.ProviderStore
	hierarchyStore security.TenantHierarchyStore
}

func NewContextDetailsFactory(opts...FactoryOptions) *ContextDetailsFactory {
	opt := FactoryOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &ContextDetailsFactory {
		clientStore:    opt.ClientStore,
		accountStore:   opt.AccountStore,
		tenantStore:    opt.TenantStore,
		providerStore:  opt.ProviderStore,
		hierarchyStore: opt.HierarchyStore,
	}
}

type facts struct {
	client oauth2.OAuth2Client
	account security.Account
	tenant *security.Tenant
	provider *security.Provider
}

func (f *ContextDetailsFactory) New(ctx context.Context,
	request oauth2.OAuth2Request,
	userAuth security.Authentication) (security.ContextDetails, error) {

	facts, e := f.loadFacts(ctx, request, userAuth)
	if e != nil {
		return nil, e
	}

	return f.create(ctx, facts)
}

/**********************
	Helpers
 **********************/
func (f *ContextDetailsFactory) loadFacts(ctx context.Context,
	request oauth2.OAuth2Request,
	userAuth security.Authentication) (*facts, error) {
	client, e := f.loadClient(ctx)
	if e != nil {
		return nil, e
	}

	account, e := f.loadAccount(ctx, request, userAuth)
	if e != nil {
		return nil, e
	}

	tenant, e := f.loadTenant(ctx, request, account)
	if e != nil {
		return nil, e
	}

	provider, e := f.loadProvider(ctx, request, tenant)
	if e != nil {
		return nil, e
	}

	return &facts {
		client: client,
		account: account,
		tenant: tenant,
		provider: provider,
	}, nil
}

func (f *ContextDetailsFactory) loadAccount(ctx context.Context, request oauth2.OAuth2Request, userAuth security.Authentication) (security.Account, error) {
	// sanity check, this should not happen
	if userAuth.State() < security.StateAuthenticated || userAuth.Principal() == nil {
		return nil, oauth2.NewInternalError("trying create context details with unauthenticated user")
	}

	// we want to reload user's account
	principal := userAuth.Principal()
	var username string
	switch principal.(type) {
	case security.Account:
		username = principal.(security.Account).Username()
	case string:
		username = principal.(string)
	case fmt.Stringer:
		username = principal.(fmt.Stringer).String()
	default:
		return nil, oauth2.NewInternalError(fmt.Sprintf("unsupported principal type %T", principal))
	}

	acct, e := f.accountStore.LoadAccountByUsername(ctx, username)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError("invalid authorizing user", e)
	}
	return acct, nil
}

func (f *ContextDetailsFactory) loadTenant(ctx context.Context, request oauth2.OAuth2Request, account security.Account) (*security.Tenant, error) {
	tenancy, ok := account.(security.AccountTenancy)
	if !ok {
		return nil, oauth2.NewInvalidGrantError(fmt.Sprintf("account [%T] does not provide tenancy information", account))
	}

	// extract tenant id or name
	tenantId, idOk := request.Parameters()[oauth2.ParameterTenantId]
	tenantName, nOk := request.Parameters()[oauth2.ParameterTenantName]
	if (!idOk || tenantId == "") && (!nOk || tenantName == "") {
		tenantId = tenancy.DefaultTenantId()
	}

	var tenant *security.Tenant
	var e error
	if tenantId != "" {
		tenant, e = f.tenantStore.LoadTenantById(ctx, tenantId)
		if e != nil {
			return nil, oauth2.NewInvalidGrantError(fmt.Sprintf("user [%s] does not access tenant with id [%s]", account.Username(), tenantId))
		}
	} else {
		tenant, e = f.tenantStore.LoadTenantByName(ctx, tenantName)
		if e != nil {
			return nil, oauth2.NewInvalidGrantError(fmt.Sprintf("user [%s] does not access tenant with name [%s]", account.Username(), tenantName))
		}
	}

	// TODO maybe check tenant access here (both client and user)
	return tenant, nil
}

func (f *ContextDetailsFactory) loadProvider(ctx context.Context, request oauth2.OAuth2Request, tenant *security.Tenant) (*security.Provider, error) {
	providerId := tenant.ProviderId
	if providerId == "" {
		return nil, oauth2.NewInvalidGrantError("provider ID is not avalilable")
	}

	provider, e := f.providerStore.LoadProviderById(ctx, providerId)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(fmt.Sprintf("tenant [%s]'s provider is invalid", tenant.Name))
	}
	return provider, nil
}

func (f *ContextDetailsFactory) loadClient(ctx context.Context) (oauth2.OAuth2Client, error) {
	clientAuth := security.Get(ctx)
	// sanity check, this should not happen
	if clientAuth.State() < security.StateAuthenticated || clientAuth.Principal() == nil {
		return nil, oauth2.NewInternalError("trying create context details with unknown client")
	}

	// we don't typically reload client because client is usually authenticated within same transaction
	principal := clientAuth.Principal()
	switch principal.(type) {
	case oauth2.OAuth2Client:
		return principal.(oauth2.OAuth2Client), nil
	case string:
		clientId := principal.(string)
		return f.clientStore.LoadClientByClientId(ctx, clientId)
	case fmt.Stringer:
		clientId := principal.(fmt.Stringer).String()
		return f.clientStore.LoadClientByClientId(ctx, clientId)
	default:
		return nil, oauth2.NewInternalError(fmt.Sprintf("unsupported client principal type %T", principal))
	}
}

func (f *ContextDetailsFactory) create(ctx context.Context, facts *facts) (security.ContextDetails, error) {
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
		LocaleCode: "", // TODO
		CurrencyCode: "",  // TODO
		FirstName: "", // TODO
		LastName: "", // TODO
		Email: "", // TODO
	}

	// creds
	ad, e := f.createAuthDetails(ctx, facts) // TODO
	if e != nil {
		return nil, e
	}

	return &internal.ContextDetails{
		Provider: pd,
		Tenant: td,
		User: ud,
		Authentication: *ad,
		KV: map[string]interface{}{},
	},  nil
}

func (f *ContextDetailsFactory) createAuthDetails(ctx context.Context, facts *facts) (*internal.AuthenticationDetails, error) {
	d := internal.AuthenticationDetails {
		IssueTime: time.Time{}, // TODO
		ExpiryTime: time.Time{},// TODO
		Roles: nil,// TODO
		Permissions: utils.NewStringSet(facts.account.Permissions()...),
		AuthenticationTime: time.Time{}, // TODO
		OriginalUsername: "",
		Proxied: false,
	}
	return &d, nil
}