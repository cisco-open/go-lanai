package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
	"time"
)

type UserDetails struct {
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
	UserDetails
	LockingRule
	PasswordPolicy
}

func NewUsernamePasswordAccount(details *UserDetails) *UsernamePasswordAccount {
	return &UsernamePasswordAccount{ UserDetails: *details}
}

/***********************************
	implements security.Account
 ***********************************/
func (a *UsernamePasswordAccount) ID() interface{} {
	return a.UserDetails.ID
}

func (a *UsernamePasswordAccount) Type() security.AccountType {
	return a.UserDetails.Type
}

func (a *UsernamePasswordAccount) Username() string {
	return a.UserDetails.Username
}

func (a *UsernamePasswordAccount) Credentials() interface{} {
	return a.UserDetails.Credentials
}

func (a *UsernamePasswordAccount) Permissions() []string {
	return a.UserDetails.Permissions
}

func (a *UsernamePasswordAccount) Disabled() bool {
	return a.UserDetails.Disabled
}

func (a *UsernamePasswordAccount) Locked() bool {
	return a.UserDetails.Locked
}

func (a *UsernamePasswordAccount) UseMFA() bool {
	return a.UserDetails.UseMFA
}

/***********************************
	implements security.AccountTenancy
 ***********************************/
func (a *UsernamePasswordAccount) DefaultTenantId() string {
	return a.UserDetails.DefaultTenantId
}

func (a *UsernamePasswordAccount) Tenants() []string {
	return a.UserDetails.Tenants
}

/***********************************
	implements security.AccountHistory
 ***********************************/
func (a *UsernamePasswordAccount) LastLoginTime() time.Time {
	return a.UserDetails.LastLoginTime
}

func (a *UsernamePasswordAccount) LoginFailures() []time.Time {
	return a.UserDetails.LoginFailures
}

func (a *UsernamePasswordAccount) SerialFailedAttempts() int {
	return a.UserDetails.SerialFailedAttempts
}

func (a *UsernamePasswordAccount) PwdChangedTime() time.Time {
	return a.UserDetails.PwdChangedTime
}

func (a *UsernamePasswordAccount) GracefulAuthCount() int {
	return a.UserDetails.GracefulAuthCount
}

/***********************************
	security.AccountUpdater
 ***********************************/
func (a *UsernamePasswordAccount) IncrementGracefulAuthCount() {
	a.UserDetails.GracefulAuthCount ++
}

func (a *UsernamePasswordAccount) Lock() {
	if !a.UserDetails.Locked {
		a.UserDetails.LockoutTime = time.Now()
	}
	a.UserDetails.Locked = true
	// TODO proper logging
	fmt.Printf("Account[%s] Locked\n", a.UserDetails.Username)
}

func (a *UsernamePasswordAccount) Unlock() {
	// we don't clear lockout time for record keeping purpose
	a.UserDetails.Locked = false
	// TODO proper logging
	fmt.Printf("Account[%s] Unlocked\n", a.UserDetails.Username)
}

func (a *UsernamePasswordAccount) RecordFailure(failureTime time.Time, limit int) {
	failures := append(a.UserDetails.LoginFailures, failureTime)
	if len(failures) > limit {
		failures = failures[len(failures) - limit:]
	}
	a.UserDetails.LoginFailures = failures
	a.UserDetails.SerialFailedAttempts = len(failures)
}

func (a *UsernamePasswordAccount) RecordSuccess(loginTime time.Time) {
	a.UserDetails.LastLoginTime = loginTime
}

func (a *UsernamePasswordAccount) ResetFailedAttempts() {
	a.UserDetails.SerialFailedAttempts = 0
	a.UserDetails.LoginFailures = []time.Time{}
	// TODO proper logging
	fmt.Printf("Account[%s] Failure reset\n", a.UserDetails.Username)
}

func (a *UsernamePasswordAccount) ResetGracefulAuthCount() {
	a.UserDetails.GracefulAuthCount = 0
	// TODO proper logging
	fmt.Printf("Account[%s] Graceful Auth Reset\n", a.UserDetails.Username)
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
	return a.UserDetails.LockoutTime
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
