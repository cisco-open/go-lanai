package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
)

type OAuth2AccountStore struct {
	accountStore security.AccountStore
	clientStore  oauth2.OAuth2ClientStore
}

func (o *OAuth2AccountStore) Finalize(ctx context.Context, account security.Account, options ...security.AccountFinalizeOptions) (security.Account, error) {
	if f, ok := o.accountStore.(security.AccountFinalizer); ok {
		ddt := account.(security.AccountTenancy).DefaultDesignatedTenantId()
		dt := account.(security.AccountTenancy).DesignatedTenantIds()
		newAcct, err := f.Finalize(ctx, account, options...)
		if err != nil {
			return nil, err
		} else {
			return &OAuth2Account{
				account:                   newAcct,
				designatedTenantIds:       dt,
				defaultDesignatedTenantId: ddt,
			}, nil
		}
	} else {
		return account, nil
	}
}

func NewOAuth2AccountStore(a security.AccountStore, c oauth2.OAuth2ClientStore) *OAuth2AccountStore {
	return &OAuth2AccountStore{
		accountStore: a,
		clientStore:  c,
	}
}

func (o *OAuth2AccountStore) LoadAccountById(ctx context.Context, id interface{}, clientId string) (security.Account, error) {
	c, err := o.clientStore.LoadClientByClientId(ctx, clientId)

	if err != nil {
		return nil, err
	}

	a, err := o.accountStore.LoadAccountById(ctx, id)

	if err != nil {
		return nil, err
	}

	if _, ok := a.(security.AccountTenancy); !ok {
		return nil, errors.New("account must have tenancy")
	}

	var effectiveTenants []string

	for _, t := range a.(security.AccountTenancy).DesignatedTenantIds() {
		if c.Scopes().Has(oauth2.ScopeSystem) || tenancy.AnyHasDescendant(ctx, c.AssignedTenantIds(), t) {
			effectiveTenants = append(effectiveTenants, t)
		}
	}

	var defaultTenantId string
	if tenancy.AnyHasDescendant(ctx,
		utils.NewStringSet(effectiveTenants...),
		a.(security.AccountTenancy).DefaultDesignatedTenantId()) {
		defaultTenantId = a.(security.AccountTenancy).DefaultDesignatedTenantId()
	}

	return &OAuth2Account{
		account:                   a,
		designatedTenantIds:       effectiveTenants,
		defaultDesignatedTenantId: defaultTenantId,
	}, nil
}

func (o *OAuth2AccountStore) LoadAccountByUsername(ctx context.Context, username string, clientId string) (security.Account, error) {
	c, err := o.clientStore.LoadClientByClientId(ctx, clientId)

	if err != nil {
		return nil, err
	}

	a, err := o.accountStore.LoadAccountByUsername(ctx, username)

	if err != nil {
		return nil, err
	}

	if _, ok := a.(security.AccountTenancy); !ok {
		return nil, errors.New("account must have tenancy")
	}

	effectiveTenants := utils.NewStringSet()

	for _, t := range a.(security.AccountTenancy).DesignatedTenantIds() {
		// if user's tenant is children of client's tenant, add it
		if c.Scopes().Has(oauth2.ScopeSystem) || tenancy.AnyHasDescendant(ctx, c.AssignedTenantIds(), t) {
			effectiveTenants.Add(t)
		}
	}

	userPerms := utils.NewStringSet(a.Permissions()...)
	for t, _ := range c.AssignedTenantIds() {
		if userPerms.Has(security.SpecialPermissionAccessAllTenant) || tenancy.AnyHasDescendant(ctx,
			utils.NewStringSet(a.(security.AccountTenancy).DesignatedTenantIds()...), t) {
			effectiveTenants.Add(t)
		}
	}

	var defaultTenantId string
	if tenancy.AnyHasDescendant(ctx,
		effectiveTenants,
		a.(security.AccountTenancy).DefaultDesignatedTenantId()) {
		defaultTenantId = a.(security.AccountTenancy).DefaultDesignatedTenantId()
	}

	return &OAuth2Account{
		account:                   a,
		designatedTenantIds:       effectiveTenants.Values(),
		defaultDesignatedTenantId: defaultTenantId,
	}, nil
}

type OAuth2Account struct {
	account security.Account

	defaultDesignatedTenantId string
	designatedTenantIds       []string
}

/***********************************
	implements security.Account
 ***********************************/

func (a *OAuth2Account) ID() interface{} {
	return a.account.ID()
}

func (a *OAuth2Account) Type() security.AccountType {
	return a.account.Type()
}

func (a *OAuth2Account) Username() string {
	return a.account.Username()
}

func (a *OAuth2Account) Credentials() interface{} {
	return a.account.Credentials()
}

func (a *OAuth2Account) Permissions() []string {
	return a.account.Permissions()
}

func (a *OAuth2Account) Disabled() bool {
	return a.account.Disabled()
}

func (a *OAuth2Account) Locked() bool {
	return a.account.Locked()
}

func (a *OAuth2Account) UseMFA() bool {
	return a.account.UseMFA()
}

func (a *OAuth2Account) CacheableCopy() security.Account {
	return a.account.CacheableCopy()
}

/***********************************
	implements security.AccountTenancy
 ***********************************/

func (a *OAuth2Account) DefaultDesignatedTenantId() string {
	return a.defaultDesignatedTenantId
}

func (a *OAuth2Account) DesignatedTenantIds() []string {
	return a.designatedTenantIds
}

func (a *OAuth2Account) TenantId() string {
	return a.account.(security.AccountTenancy).TenantId()
}

/***********************************
	security.AcctMetadata
 ***********************************/

func (a *OAuth2Account) RoleNames() []string {
	if m, ok := a.account.(security.AccountMetadata); ok {
		return m.RoleNames()
	} else {
		return nil
	}
}

func (a *OAuth2Account) FirstName() string {
	if m, ok := a.account.(security.AccountMetadata); ok {
		return m.FirstName()
	} else {
		return ""
	}
}

func (a *OAuth2Account) LastName() string {
	if m, ok := a.account.(security.AccountMetadata); ok {
		return m.LastName()
	} else {
		return ""
	}
}

func (a *OAuth2Account) Email() string {
	if m, ok := a.account.(security.AccountMetadata); ok {
		return m.Email()
	} else {
		return ""
	}
}

func (a *OAuth2Account) LocaleCode() string {
	if m, ok := a.account.(security.AccountMetadata); ok {
		return m.LocaleCode()
	} else {
		return ""
	}
}

func (a *OAuth2Account) CurrencyCode() string {
	if m, ok := a.account.(security.AccountMetadata); ok {
		return m.CurrencyCode()
	} else {
		return ""
	}
}

func (a *OAuth2Account) Value(key string) interface{} {
	if m, ok := a.account.(security.AccountMetadata); ok {
		return m.Value(key)
	} else {
		return nil
	}
}
