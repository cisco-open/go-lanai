package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
)

type OAuth2Account struct {
	account security.Account

	defaultDesignatedTenantId string
	designatedTenantIds       []string
}

func WrapAccount(ctx context.Context, a security.Account, c oauth2.OAuth2Client) (security.Account, error) {
	if _, ok := a.(security.AccountTenancy); !ok {
		return nil, errors.New("account must have tenancy")
	}

	var effectiveTenants []string

	for _, t := range a.(security.AccountTenancy).DesignatedTenantIds() {
		if c.Scopes().Has(oauth2.ScopeCrossTenant) || tenancy.AnyHasDescendant(ctx, c.AssignedTenantIds(), t) {
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
