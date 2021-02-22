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

type FederatedAccountStore interface {
	LoadAccountByExternalId(externalIdName string, externalIdValue string, externalIdpName string) (Account, error)
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

/*********************************
	Abstraction - Multi Tenancy
 *********************************/
type AccountTenancy interface {
	DefaultTenantId() string
	TenantIds() []string
}

/*********************************
	Abstraction - Mutator
 *********************************/
type AccountUpdater interface {
	IncrementGracefulAuthCount()
	ResetGracefulAuthCount()
	Lock()
	Unlock()
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
	// PwdPolicyName the name of password polcy
	PwdAgingPolicyName() string
	// PwdPolicyEnforced indicate whether password policy is enabled
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
