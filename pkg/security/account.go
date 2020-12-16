package security

import (
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
	Type() AccountType
	Username() string
	Credentials() interface{}
	Permissions() []string
	Disabled() bool
	Locked() bool
	UseMFA() bool
}

type AccountStore interface {
	LoadAccountByUsername(username string) (Account, error)
}

/******************************
	Abstraction - Status
 ******************************/

/*********************************
	Abstraction - Auth History
 *********************************/
type AccountHistory interface {
	Account
	LastLoginTime() time.Time
	LoginFailures() []time.Time
	SerialFailedAttempts() int
}

/*********************************
	Abstraction - Locking Rules
 *********************************/
type AccountLockingRule interface {
	Account
	LockoutTime() time.Time
	LockingRuleName() string // TODO
}

/*********************************
	Abstraction - Passwd Policy
 *********************************/
type AccountPasswordPolicy interface {
	Account
	PwdChangedTime() time.Time
	PwdPolicyName() string // TODO
	GracefulAuthCount() int
}

/*********************************
	Abstraction - Multi Tenancy
 *********************************/
type AccountTenancy interface {
	Account
	DefaultTenantId() string
	Tenants() []string
}

/*********************************
	Abstraction - Mutator
 *********************************/
type AccountUpdater interface {
	Account
	IncrementFailedAttempts()
	IncrementGracefulAuthCount()
	Lock()
	Unlock()
	RecordFailure(failureTime time.Time, limit int)
	RecordSuccess(loginTime time.Time)
	ResetFailedAttempts()
	ResetGracefulAuthCount()
}