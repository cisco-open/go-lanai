package example

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"strings"
	"time"
)

var (
	startupTime             = time.Now()
	contextKeyAccountPolicy = "Policy-Acct-"
)

// in memory security.AccountStore
type InMemoryAccountStore struct {
	accountLookupByUsername map[string]*security.DefaultAccount
	accountLookupById       map[interface{}]*security.DefaultAccount
	policiesLookupByName    map[string]*PropertiesBasedAccountPolicy
}

func NewInMemoryAccountStore(acctProps AccountsProperties, acctPolicyProps AccountPoliciesProperties) security.AccountStore {
	store := InMemoryAccountStore{
		accountLookupByUsername: map[string]*security.DefaultAccount{},
		accountLookupById:       map[interface{}]*security.DefaultAccount{},
		policiesLookupByName:    map[string]*PropertiesBasedAccountPolicy{},
	}
	for _, props := range acctProps.Accounts {
		//props.ID = k
		acct := createAccount(&props)
		store.accountLookupByUsername[props.Username] = acct
		store.accountLookupById[props.ID] = acct
	}

	for k, policy := range acctPolicyProps.Policies {
		policy.Name = k
		copy := policy
		store.policiesLookupByName[policy.Name] = &copy
	}

	return &store
}

func (store *InMemoryAccountStore) Save(ctx context.Context, acct security.Account) error {
	if userAcct, ok := acct.(*security.DefaultAccount); ok {
		store.accountLookupById[userAcct.ID()] = userAcct
		store.accountLookupByUsername[userAcct.Username()] = userAcct
	}
	return nil
}

func (store *InMemoryAccountStore) LoadAccountById(_ context.Context, id interface{}) (security.Account, error) {
	u, ok := store.accountLookupById[id]
	if !ok {
		return nil, errors.New("user Domain not found")
	}

	return u, nil
}

func (store *InMemoryAccountStore) LoadAccountByUsername(_ context.Context, username string) (security.Account, error) {
	u, ok := store.accountLookupByUsername[username]
	if !ok {
		return nil, errors.New("username not found")
	}

	return u, nil
}

func (store *InMemoryAccountStore) LoadLockingRules(ctx context.Context, acct security.Account) (security.AccountLockingRule, error) {
	account, ok := acct.(*security.DefaultAccount)
	if !ok {
		return nil, errors.New("unsupported account")
	}

	if account.AcctDetails.PolicyName == "" {
		// no policy name available, return without loading (this will contains default values)
		return account, nil
	}

	policy, err := store.tryLoadPolicy(ctx, account)
	if err != nil {
		return nil, err
	}

	populateAccountPolicy(account, policy)
	return account, nil
}

func (store *InMemoryAccountStore) LoadPwdAgingRules(ctx context.Context, acct security.Account) (security.AccountPwdAgingRule, error) {
	ret, err := store.LoadLockingRules(ctx, acct)
	if err != nil {
		return nil, err
	}

	return ret.(security.AccountPwdAgingRule), nil
}

// Note, caching loaded policy in ctx is not needed for in-memory store. The implmenetation is for reference only
func (store *InMemoryAccountStore) tryLoadPolicy(ctx context.Context, account *security.DefaultAccount) (*PropertiesBasedAccountPolicy, error) {
	ctxKey := contextKeyAccountPolicy + account.ID().(string)
	policy, ok := ctx.Value(ctxKey).(*PropertiesBasedAccountPolicy)
	if !ok {
		// load it
		policy, ok = store.policiesLookupByName[account.PolicyName]
	}

	if !ok {
		return nil, errors.New("account policy not found")
	}

	// try to cache loaded policy in context
	if mutable, ok := ctx.(utils.MutableContext); ok {
		mutable.Set(ctxKey, policy)
	}
	return policy, nil
}

func createAccount(props *PropertiesBasedAccount) *security.DefaultAccount {
	acct := security.NewUsernamePasswordAccount(&security.AcctDetails{
		ID:              props.ID,
		Type:            security.ParseAccountType(props.Type),
		Username:        props.Username,
		Credentials:     props.Password,
		Permissions:     props.Permissions,
		Disabled:        props.Disabled,
		Locked:          props.Locked,
		UseMFA:          props.UseMFA,
		DefaultTenantId: props.DefaultTenantId,
		Tenants:         props.Tenants,
		PolicyName:      props.AccountPolicyName,

		LastLoginTime: startupTime.Add(-2 * time.Hour),
		LoginFailures: []time.Time {
			startupTime.Add(-115 * time.Minute), // 60 mins till next failure
			startupTime.Add(-55 * time.Minute),  // 30 mins till next failure
			startupTime.Add(-25 * time.Minute),  // 15 mins till next failure
			startupTime.Add(-10 * time.Minute),  // 8 mins till next failure
			startupTime.Add(-2 * time.Minute),
		},
		SerialFailedAttempts: 5,
		LockoutTime:          startupTime.Add(-30 * time.Minute),
		PwdChangedTime:       startupTime.Add(-30 * 24 * time.Hour),
		GracefulAuthCount:    0,
	})

	// metadata
	names := strings.Split(props.FullName, " ")
	if len(names) > 0 {
		acct.AcctMetadata.FirstName = names[0]
	}

	if len(names) > 1 {
		acct.AcctMetadata.LastName = names[len(names) - 1]
	}
	acct.AcctMetadata.Email = props.Email
	acct.AcctMetadata.RoleNames = props.Permissions
	acct.AcctMetadata.Extra = map[string]interface{}{}

	return acct
}

func populateAccountPolicy(acct *security.DefaultAccount, props *PropertiesBasedAccountPolicy) {
	acct.AcctLockingRule = security.AcctLockingRule{
		Name:             props.Name,
		Enabled:          props.LockingEnabled,
		LockoutDuration:  utils.ParseDuration(props.LockoutDuration),
		FailuresLimit:    props.FailuresLimit,
		FailuresInterval: utils.ParseDuration(props.FailuresInterval),
	}

	acct.AcctPasswordPolicy = security.AcctPasswordPolicy{
		Name:                props.Name,
		Enabled:             props.AgingEnabled,
		MaxAge:              utils.ParseDuration(props.MaxAge),
		ExpiryWarningPeriod: utils.ParseDuration(props.ExpiryWarningPeriod),
		GracefulAuthLimit:   props.GracefulAuthLimit,
	}
}

type InMemoryFederatedAccountStore struct {
	
}

func (i *InMemoryFederatedAccountStore) LoadAccountByExternalId(externalIdName string, externalIdValue string, externalIdpName string) (security.Account, error) {
	return security.NewUsernamePasswordAccount(&security.AcctDetails{
		ID:              "user-tishi",
		Type:            security.AccountTypeFederated,
		Username:        "tishi",
		Permissions: []string{"welcomed"}}), nil
}

func NewInMemoryFederatedAccountStore() security.FederatedAccountStore{
	return &InMemoryFederatedAccountStore{}
}