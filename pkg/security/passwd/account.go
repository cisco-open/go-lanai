package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
	"time"
)

type AccountDetails struct {
	ID 					 string
	Type                 security.AccountType
	Username             string
	Credentials          interface{}
	Permissions          []string
	Disabled             bool
	Locked               bool
	UseMFA               bool
	DefaultTenantId      string
	Tenants              []string
	LastLoginTime        time.Time
	LoginFailures        []time.Time
	SerialFailedAttempts int
	LockoutTime          time.Time
	PwdChangedTime       time.Time
	GracefulAuthCount    int
	PolicyName           string
}

type LockingRule struct {
	Name             string
	Enabled          bool
	LockoutDuration  time.Duration
	FailuresLimit    int
	FailuresInterval time.Duration
}

type PasswordPolicy struct {
	Name                string
	Enabled             bool
	MaxAge              time.Duration
	ExpiryWarningPeriod time.Duration
	GracefulAuthLimit   int
}

type UsernamePasswordAccount struct {
	AccountDetails
	LockingRule
	PasswordPolicy
}

func NewUsernamePasswordAccount(details *AccountDetails) *UsernamePasswordAccount {
	return &UsernamePasswordAccount{ AccountDetails: *details}
}

/***********************************
	implements security.Account
 ***********************************/
func (a *UsernamePasswordAccount) ID() interface{} {
	return a.AccountDetails.ID
}

func (a *UsernamePasswordAccount) Type() security.AccountType {
	return a.AccountDetails.Type
}

func (a *UsernamePasswordAccount) Username() string {
	return a.AccountDetails.Username
}

func (a *UsernamePasswordAccount) Credentials() interface{} {
	return a.AccountDetails.Credentials
}

func (a *UsernamePasswordAccount) Permissions() []string {
	return a.AccountDetails.Permissions
}

func (a *UsernamePasswordAccount) Disabled() bool {
	return a.AccountDetails.Disabled
}

func (a *UsernamePasswordAccount) Locked() bool {
	return a.AccountDetails.Locked
}

func (a *UsernamePasswordAccount) UseMFA() bool {
	return a.AccountDetails.UseMFA
}

/***********************************
	implements security.AccountTenancy
 ***********************************/
func (a *UsernamePasswordAccount) DefaultTenantId() string {
	return a.AccountDetails.DefaultTenantId
}

func (a *UsernamePasswordAccount) Tenants() []string {
	return a.AccountDetails.Tenants
}

/***********************************
	implements security.AccountHistory
 ***********************************/
func (a *UsernamePasswordAccount) LastLoginTime() time.Time {
	return a.AccountDetails.LastLoginTime
}

func (a *UsernamePasswordAccount) LoginFailures() []time.Time {
	return a.AccountDetails.LoginFailures
}

func (a *UsernamePasswordAccount) SerialFailedAttempts() int {
	return a.AccountDetails.SerialFailedAttempts
}

func (a *UsernamePasswordAccount) PwdChangedTime() time.Time {
	return a.AccountDetails.PwdChangedTime
}

func (a *UsernamePasswordAccount) GracefulAuthCount() int {
	return a.AccountDetails.GracefulAuthCount
}

/***********************************
	security.AccountUpdater
 ***********************************/
func (a *UsernamePasswordAccount) IncrementGracefulAuthCount() {
	a.AccountDetails.GracefulAuthCount ++
}

func (a *UsernamePasswordAccount) Lock() {
	if !a.AccountDetails.Locked {
		a.AccountDetails.LockoutTime = time.Now()
	}
	a.AccountDetails.Locked = true
	// TODO proper logging
	fmt.Printf("Account[%s] Locked\n", a.AccountDetails.Username)
}

func (a *UsernamePasswordAccount) Unlock() {
	// we don't clear lockout time for record keeping purpose
	a.AccountDetails.Locked = false
	// TODO proper logging
	fmt.Printf("Account[%s] Unlocked\n", a.AccountDetails.Username)
}

func (a *UsernamePasswordAccount) RecordFailure(failureTime time.Time, limit int) {
	failures := append(a.AccountDetails.LoginFailures, failureTime)
	if len(failures) > limit {
		failures = failures[len(failures) - limit:]
	}
	a.AccountDetails.LoginFailures = failures
	a.AccountDetails.SerialFailedAttempts = len(failures)
}

func (a *UsernamePasswordAccount) RecordSuccess(loginTime time.Time) {
	a.AccountDetails.LastLoginTime = loginTime
}

func (a *UsernamePasswordAccount) ResetFailedAttempts() {
	a.AccountDetails.SerialFailedAttempts = 0
	a.AccountDetails.LoginFailures = []time.Time{}
	// TODO proper logging
	fmt.Printf("Account[%s] Failure reset\n", a.AccountDetails.Username)
}

func (a *UsernamePasswordAccount) ResetGracefulAuthCount() {
	a.AccountDetails.GracefulAuthCount = 0
	// TODO proper logging
	fmt.Printf("Account[%s] Graceful Auth Reset\n", a.AccountDetails.Username)
}

/***********************************
	security.AccountLockingRule
 ***********************************/
func (a *UsernamePasswordAccount) LockoutPolicyName() string {
	return a.LockingRule.Name
}

func (a *UsernamePasswordAccount) LockoutEnabled() bool {
	return a.LockingRule.Enabled
}

func (a *UsernamePasswordAccount) LockoutTime() time.Time {
	return a.AccountDetails.LockoutTime
}

func (a *UsernamePasswordAccount) LockoutDuration() time.Duration {
	return a.LockingRule.LockoutDuration
}

func (a *UsernamePasswordAccount) LockoutFailuresLimit() int {
	return a.LockingRule.FailuresLimit
}

func (a *UsernamePasswordAccount) LockoutFailuresInterval() time.Duration {
	return a.LockingRule.FailuresInterval
}

/***********************************
	security.AccountPwdAgingRule
 ***********************************/
func (a *UsernamePasswordAccount) PwdAgingPolicyName() string {
	return a.PasswordPolicy.Name
}

func (a *UsernamePasswordAccount) PwdAgingRuleEnforced() bool {
	return a.PasswordPolicy.Enabled
}

func (a *UsernamePasswordAccount) PwdMaxAge() time.Duration {
	return a.PasswordPolicy.MaxAge
}

func (a *UsernamePasswordAccount) PwdExpiryWarningPeriod() time.Duration {
	return a.PasswordPolicy.ExpiryWarningPeriod
}

func (a *UsernamePasswordAccount) GracefulAuthLimit() int {
	return a.PasswordPolicy.GracefulAuthLimit
}
