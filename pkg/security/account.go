package security

import (
	"context"
	"strings"
	"time"
)

/******************************
	Abstraction - Basics
 ******************************/

type AccountType int

const (
	AccountTypeUnknown AccountType = iota
	AccountTypeDefault
	AccountTypeApp
	AccountTypeFederated
	AccountTypeSystem
)

func (t AccountType) String() string {
	switch t {
	case AccountTypeDefault:
		return "user"
	case AccountTypeApp:
		return "app"
	case AccountTypeFederated:
		return "fed"
	case AccountTypeSystem:
		return "system"
	default:
		return ""
	}
}

func ParseAccountType(value interface{}) AccountType {
	if v, ok := value.(AccountType); ok {
		return v
	}

	switch v, ok := value.(string); ok {
	case "user" == strings.ToLower(v):
		return AccountTypeDefault
	case "app" == strings.ToLower(v):
		return AccountTypeApp
	case "fed" == strings.ToLower(v):
		return AccountTypeFederated
	case "system" == strings.ToLower(v):
		return AccountTypeSystem
	default:
		return AccountTypeUnknown
	}
}

type Account interface {
	ID() interface{}
	Type() AccountType
	Username() string
	Credentials() interface{}
	Permissions() []string
	Disabled() bool
	Locked() bool
	UseMFA() bool
	// CacheableCopy should returns a copy of Account that suitable for putting into cache.
	// e.g. the CacheableCopy should be able to be serialized and shouldn't contains Credentials or any reloadable content
	CacheableCopy() Account
}

type AccountFinalizeOption struct {
	Tenant *Tenant
}
type AccountFinalizeOptions func(option *AccountFinalizeOption)

func FinalizeWithTenant(tenant *Tenant) AccountFinalizeOptions {
	return func(option *AccountFinalizeOption) {
		option.Tenant = tenant
	}
}

type AccountFinalizer interface {
	// Finalize is a function that will allow a service to modify the account before it
	// is put into the security context. An example usage of this is to allow for per-tenant
	// permissions where a user can have different permissions depending on which tenant is selected.
	//
	// Note that the Account.ID and Account.Username should not be changed. If those fields are changed
	// an error will be reported.
	Finalize(ctx context.Context, account Account, options ...AccountFinalizeOptions) (Account, error)
}

type AccountStore interface {
	// LoadAccountById find account by its Domain
	LoadAccountById(ctx context.Context, id interface{}) (Account, error)
	// LoadAccountByUsername find account by its Username
	LoadAccountByUsername(ctx context.Context, username string) (Account, error)
	// LoadLockingRules load given account's locking rule. It's recommended to cache the result
	LoadLockingRules(ctx context.Context, acct Account) (AccountLockingRule, error)
	// LoadPwdAgingRules load given account's password policy. It's recommended to cache the result
	LoadPwdAgingRules(ctx context.Context, acct Account) (AccountPwdAgingRule, error)
	// Save save the account if necessary
	Save(ctx context.Context, acct Account) error
}

type AutoCreateUserDetails interface {
	IsEnabled() bool
	GetEmailWhiteList() []string
	GetAttributeMapping() map[string]string
	GetElevatedUserRoleNames() []string
	GetRegularUserRoleNames() []string
}

type FederatedAccountStore interface {
	LoadAccountByExternalId(ctx context.Context, externalIdName string, externalIdValue string, externalIdpName string, autoCreateUserDetails AutoCreateUserDetails, rawAssertion interface{}) (Account, error)
}

/*********************************
	Abstraction - Auth History
 *********************************/

type AccountHistory interface {
	LastLoginTime() time.Time
	LoginFailures() []time.Time
	SerialFailedAttempts() int
	LockoutTime() time.Time
	PwdChangedTime() time.Time
	GracefulAuthCount() int
}

/********************************
		Abstraction - Multi Tenancy
*********************************/

type AccountTenancy interface {
	DefaultDesignatedTenantId() string
	DesignatedTenantIds() []string
	TenantId() string
}

/*********************************
	Abstraction - Mutator
 *********************************/

type AccountUpdater interface {
	IncrementGracefulAuthCount()
	ResetGracefulAuthCount()
	LockAccount()
	UnlockAccount()
	RecordFailure(failureTime time.Time, limit int)
	RecordSuccess(loginTime time.Time)
	ResetFailedAttempts()
}

/*********************************
	Abstraction - Locking Rules
 *********************************/

type AccountLockingRule interface {
	// LockoutPolicyName the name of locking rule
	LockoutPolicyName() string
	// LockoutEnabled indicate whether account locking is enabled
	LockoutEnabled() bool
	// LockoutDuration specify how long the account should be locked after consecutive login failures
	LockoutDuration() time.Duration
	// LockoutFailuresLimit specify how many consecutive login failures required to lock the account
	LockoutFailuresLimit() int
	// LockoutFailuresInterval specify how long between the first and the last login failures to be considered as consecutive login failures
	LockoutFailuresInterval() time.Duration
}

/*********************************
	Abstraction - Aging Rules
 *********************************/

type AccountPwdAgingRule interface {
	// PwdAgingPolicyName the name of password polcy
	PwdAgingPolicyName() string
	// PwdAgingRuleEnforced indicate whether password policy is enabled
	PwdAgingRuleEnforced() bool
	// PwdMaxAge specify how long a password is valid before expiry
	PwdMaxAge() time.Duration
	// PwdExpiryWarningPeriod specify how long before password expiry the system should warn user
	PwdExpiryWarningPeriod() time.Duration
	// GracefulAuthLimit specify how many logins is allowed after password expiry
	GracefulAuthLimit() int
}

/*********************************
	Abstraction - Metadata
 *********************************/

type AccountMetadata interface {
	RoleNames() []string
	FirstName() string
	LastName() string
	Email() string
	LocaleCode() string
	CurrencyCode() string
	Value(key string) interface{}
}
