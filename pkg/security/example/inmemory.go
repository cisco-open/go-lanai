package example

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"time"
)

var (
	startupTime             = time.Now()
	contextKeyAccountPolicy = "Policy-Acct-"
)

// in memory Store
type InMemoryStore struct {
	accountLookupByUsername map[string]*PropertiesBasedAccount
	accountLookupById       map[interface{}]*PropertiesBasedAccount
	policiesLookupByName    map[string]*PropertiesBasedAccountPolicy
}

func NewInMemoryStore(acctProps AccountsProperties, acctPolicyProps AccountPoliciesProperties) security.AccountStore {
	store := InMemoryStore{
		accountLookupByUsername: map[string]*PropertiesBasedAccount{},
		accountLookupById:       map[interface{}]*PropertiesBasedAccount{},
		policiesLookupByName:    map[string]*PropertiesBasedAccountPolicy{},
	}
	for k,acct := range acctProps.Accounts {
		acct.ID = k
		copy := acct
		store.accountLookupByUsername[acct.Username] = &copy
		store.accountLookupById[acct.ID] = &copy
	}

	for k, policy := range acctPolicyProps.Policies {
		policy.Name = k
		copy := policy
		store.policiesLookupByName[policy.Name] = &copy
	}

	return &store
}

func (store *InMemoryStore) LoadAccountById(_ context.Context, id interface{}) (security.Account, error) {
	u, ok := store.accountLookupById[id]
	if !ok {
		return nil, errors.New("user ID not found")
	}

	return createAccount(u), nil
}

func (store *InMemoryStore) LoadAccountByUsername(_ context.Context, username string) (security.Account, error) {
	u, ok := store.accountLookupByUsername[username]
	if !ok {
		return nil, errors.New("username not found")
	}

	return createAccount(u), nil
}

func (store *InMemoryStore) LoadLockingRules(ctx context.Context, acct security.Account) (security.AccountLockingRule, error) {
	account, ok := acct.(*passwd.UsernamePasswordAccount)
	if !ok {
		return nil, errors.New("unsupported account")
	}

	if account.UserDetails.PolicyName == "" {
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

func (store *InMemoryStore) LoadPasswordPolicy(ctx context.Context, acct security.Account) (security.AccountPasswordPolicy, error) {
	ret, err := store.LoadLockingRules(ctx, acct)
	if err != nil {
		return nil, err
	}

	return ret.(security.AccountPasswordPolicy), nil
}

// Note, caching loaded policy in ctx is not needed for in-memory store. The implmenetation is for reference only
func (store *InMemoryStore) tryLoadPolicy(ctx context.Context, account *passwd.UsernamePasswordAccount) (*PropertiesBasedAccountPolicy, error) {
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
	utils.MakeMutableContext(ctx).SetValue(ctxKey, policy)
	return policy, nil
}

func createAccount(props *PropertiesBasedAccount) security.Account {
	acct := passwd.NewUsernamePasswordAccount(&passwd.UserDetails{
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
		LockoutTime:          startupTime.Add(-2 * time.Minute),
		PwdChangedTime:       startupTime.Add(-30 * 24 * time.Hour),
		GracefulAuthCount:    0,
	})

	return acct
}

func populateAccountPolicy(acct *passwd.UsernamePasswordAccount, props *PropertiesBasedAccountPolicy) {
	acct.LockingRule = passwd.LockingRule{
		Name:             props.Name,
		Enabled:          props.LockingEnabled,
		LockoutDuration:  utils.ParseDuration(props.LockoutDuration),
		FailuresLimit:    props.FailuresLimit,
		FailuresInterval: utils.ParseDuration(props.FailuresInterval),
	}

	acct.PasswordPolicy = passwd.PasswordPolicy{
		Name:                props.Name,
		Enabled:             props.AgingEnabled,
		MaxAge:              utils.ParseDuration(props.MaxAge),
		ExpiryWarningPeriod: utils.ParseDuration(props.ExpiryWarningPeriod),
		GracefulAuthLimit:   props.GracefulAuthLimit,
	}
}