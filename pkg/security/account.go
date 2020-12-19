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
	AccountTypeDefault AccountType = iota
	AccountTypeApp
	AccountTypeFederated
	AccountTypeUnknown
)
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
}

type AccountStore interface {
	LoadAccountById(ctx context.Context, id interface{}) (Account, error)
	LoadAccountByUsername(ctx context.Context, username string) (Account, error)
	LoadLockingRules(ctx context.Context, acct Account) (AccountLockingRule, error)
	LoadPasswordPolicy(ctx context.Context, acct Account) (AccountPasswordPolicy, error)
	Save(ctx context.Context, acct Account) error
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
	Tenants() []string
}

/*********************************
	Abstraction - Mutator
 *********************************/
type AccountUpdater interface {
	IncrementFailedAttempts()
	IncrementGracefulAuthCount()
	Lock()
	Unlock()
	RecordFailure(failureTime time.Time, limit int)
	RecordSuccess(loginTime time.Time)
	ResetFailedAttempts()
	ResetGracefulAuthCount()
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
	// LockoutFailuresInterval specify how long between two login failures to be considered as  consecutive login failures
	LockoutFailuresInterval() time.Duration
}

/*********************************
	Abstraction - Passwd Policy
 *********************************/
type AccountPasswordPolicy interface {
	// PwdPolicyName the name of password polcy
	PwdPolicyName() string
	// PwdPolicyEnforced indicate whether password policy is enabled
	PwdPolicyEnforced() bool
	// PwdMaxAge specify how long a password is valid before expiry
	PwdMaxAge() time.Duration
	// PwdExpiryWarningPeriod specify how long before password expiry the system should warn user
	PwdExpiryWarningPeriod() time.Duration
	// GracefulAuthLimit specify how many logins is allowed after password expiry
	GracefulAuthLimit() int
}