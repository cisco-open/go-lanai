// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"time"
)

type AcctDetails struct {
	ID                        string
	Type                      AccountType
	Username                  string
	Credentials               interface{}
	Permissions               []string
	Disabled                  bool
	Locked                    bool
	UseMFA                    bool
	DefaultDesignatedTenantId string
	DesignatedTenantIds       []string
	TenantId                  string
	LastLoginTime             time.Time
	LoginFailures             []time.Time
	SerialFailedAttempts      int
	LockoutTime               time.Time
	PwdChangedTime            time.Time
	GracefulAuthCount         int
	PolicyName                string
}

type AcctLockingRule struct {
	Name             string
	Enabled          bool
	LockoutDuration  time.Duration
	FailuresLimit    int
	FailuresInterval time.Duration
}

type AcctPasswordPolicy struct {
	Name                string
	Enabled             bool
	MaxAge              time.Duration
	ExpiryWarningPeriod time.Duration
	GracefulAuthLimit   int
}

type AcctMetadata struct {
	RoleNames    []string
	FirstName    string
	LastName     string
	Email        string
	LocaleCode   string
	CurrencyCode string
	Extra        map[string]interface{}
}

type DefaultAccount struct {
	AcctDetails
	AcctLockingRule
	AcctPasswordPolicy
	AcctMetadata
}

func NewUsernamePasswordAccount(details *AcctDetails) *DefaultAccount {
	return &DefaultAccount{AcctDetails: *details}
}

/***********************************
	implements security.Account
 ***********************************/

func (a *DefaultAccount) ID() interface{} {
	return a.AcctDetails.ID
}

func (a *DefaultAccount) Type() AccountType {
	return a.AcctDetails.Type
}

func (a *DefaultAccount) Username() string {
	return a.AcctDetails.Username
}

func (a *DefaultAccount) Credentials() interface{} {
	return a.AcctDetails.Credentials
}

func (a *DefaultAccount) Permissions() []string {
	return a.AcctDetails.Permissions
}

func (a *DefaultAccount) Disabled() bool {
	return a.AcctDetails.Disabled
}

func (a *DefaultAccount) Locked() bool {
	return a.AcctDetails.Locked
}

func (a *DefaultAccount) UseMFA() bool {
	return a.AcctDetails.UseMFA
}

func (a *DefaultAccount) CacheableCopy() Account {
	cp := DefaultAccount{
		AcctDetails:  a.AcctDetails,
		AcctMetadata: a.AcctMetadata,
	}
	cp.AcctDetails.Credentials = nil
	return &cp
}

/***********************************
	implements security.AccountTenancy
 ***********************************/

func (a *DefaultAccount) DefaultDesignatedTenantId() string {
	return a.AcctDetails.DefaultDesignatedTenantId
}

func (a *DefaultAccount) DesignatedTenantIds() []string {
	return a.AcctDetails.DesignatedTenantIds
}

func (a *DefaultAccount) TenantId() string {
	return a.AcctDetails.TenantId
}

/***********************************
	implements security.AccountHistory
 ***********************************/

func (a *DefaultAccount) LastLoginTime() time.Time {
	return a.AcctDetails.LastLoginTime
}

func (a *DefaultAccount) LoginFailures() []time.Time {
	return a.AcctDetails.LoginFailures
}

func (a *DefaultAccount) SerialFailedAttempts() int {
	return a.AcctDetails.SerialFailedAttempts
}

func (a *DefaultAccount) PwdChangedTime() time.Time {
	return a.AcctDetails.PwdChangedTime
}

func (a *DefaultAccount) GracefulAuthCount() int {
	return a.AcctDetails.GracefulAuthCount
}

/***********************************
	security.AccountUpdater
 ***********************************/

func (a *DefaultAccount) IncrementGracefulAuthCount() {
	a.AcctDetails.GracefulAuthCount++
}

func (a *DefaultAccount) LockAccount() {
	if !a.AcctDetails.Locked {
		a.AcctDetails.LockoutTime = time.Now()
	}
	a.AcctDetails.Locked = true
}

func (a *DefaultAccount) UnlockAccount() {
	// we don't clear lockout time for record keeping purpose
	a.AcctDetails.Locked = false
}

func (a *DefaultAccount) RecordFailure(failureTime time.Time, limit int) {
	failures := append(a.AcctDetails.LoginFailures, failureTime)
	if len(failures) > limit {
		failures = failures[len(failures)-limit:]
	}
	a.AcctDetails.LoginFailures = failures
	a.AcctDetails.SerialFailedAttempts = len(failures)
}

func (a *DefaultAccount) RecordSuccess(loginTime time.Time) {
	a.AcctDetails.LastLoginTime = loginTime
}

func (a *DefaultAccount) ResetFailedAttempts() {
	a.AcctDetails.SerialFailedAttempts = 0
	a.AcctDetails.LoginFailures = []time.Time{}
}

func (a *DefaultAccount) ResetGracefulAuthCount() {
	a.AcctDetails.GracefulAuthCount = 0
}

/***********************************
	security.AccountLockingRule
 ***********************************/

func (a *DefaultAccount) LockoutPolicyName() string {
	return a.AcctLockingRule.Name
}

func (a *DefaultAccount) LockoutEnabled() bool {
	return a.AcctLockingRule.Enabled
}

func (a *DefaultAccount) LockoutTime() time.Time {
	return a.AcctDetails.LockoutTime
}

func (a *DefaultAccount) LockoutDuration() time.Duration {
	return a.AcctLockingRule.LockoutDuration
}

func (a *DefaultAccount) LockoutFailuresLimit() int {
	return a.AcctLockingRule.FailuresLimit
}

func (a *DefaultAccount) LockoutFailuresInterval() time.Duration {
	return a.AcctLockingRule.FailuresInterval
}

/***********************************
	security.AccountPwdAgingRule
 ***********************************/

func (a *DefaultAccount) PwdAgingPolicyName() string {
	return a.AcctPasswordPolicy.Name
}

func (a *DefaultAccount) PwdAgingRuleEnforced() bool {
	return a.AcctPasswordPolicy.Enabled
}

func (a *DefaultAccount) PwdMaxAge() time.Duration {
	return a.AcctPasswordPolicy.MaxAge
}

func (a *DefaultAccount) PwdExpiryWarningPeriod() time.Duration {
	return a.AcctPasswordPolicy.ExpiryWarningPeriod
}

func (a *DefaultAccount) GracefulAuthLimit() int {
	return a.AcctPasswordPolicy.GracefulAuthLimit
}

/***********************************
	security.AcctMetadata
 ***********************************/

func (a *DefaultAccount) RoleNames() []string {
	if a.AcctMetadata.RoleNames == nil {
		return []string{}
	}
	return a.AcctMetadata.RoleNames
}

func (a *DefaultAccount) FirstName() string {
	return a.AcctMetadata.FirstName
}

func (a *DefaultAccount) LastName() string {
	return a.AcctMetadata.LastName
}

func (a *DefaultAccount) Email() string {
	return a.AcctMetadata.Email
}

func (a *DefaultAccount) LocaleCode() string {
	return a.AcctMetadata.LocaleCode
}

func (a *DefaultAccount) CurrencyCode() string {
	return a.AcctMetadata.CurrencyCode
}

func (a *DefaultAccount) Value(key string) interface{} {
	if a.AcctMetadata.Extra == nil {
		return nil
	}

	v, ok := a.AcctMetadata.Extra[key]
	if !ok {
		return nil
	}
	return v
}
