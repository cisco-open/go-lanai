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

package passwd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"errors"
	"time"
)

const (
	postProcessorOrderAccountStatus     = order.Lowest
	postProcessorOrderAccountLocking    = 0
	postProcessorOrderAdditionalDetails = order.Highest + 1
	postProcessorOrderPersistAccount    = order.Highest
)

/******************************
	abstracts
 ******************************/

// AuthenticationResult is a values carrier for PostAuthenticationProcessor
type AuthenticationResult struct {
	Candidate security.Candidate
	Auth      security.Authentication
	Error     error
}

// PostAuthenticationProcessor is invoked at the end of authentication process regardless of authentication decisions (granted or rejected)
// If PostAuthenticationProcessor implement order.Ordered interface, its order is respected using order.OrderedFirstCompareReverse.
// This means highest priority is executed last
type PostAuthenticationProcessor interface {
	// Process is invoked at the end of authentication process by the Authenticator to perform post-auth action.
	// The method is invoked regardless if the authentication is granted:
	// 	- If the authentication is granted, the AuthenticationResult.Auth is non-nil and AuthenticationResult.Error is nil
	//	- If the authentication is rejected, the AuthenticationResult.Error is non-nil and AuthenticationResult.Auth should be ignored
	//
	// If the context.Context and security.Account parameters are mutable, PostAuthenticationProcessor is allowed to change it
	// Note: PostAuthenticationProcessor typically shouldn't overwrite authentication decision (rejected or approved)
	// 		 However, it is allowed to modify result by returning different AuthenticationResult.
	//       This is useful when PostAuthenticationProcessor want to returns different error or add more details to authentication
	Process(context.Context, security.Account, AuthenticationResult) AuthenticationResult
}

/******************************
	Helpers
 ******************************/
func applyPostAuthenticationProcessors(processors []PostAuthenticationProcessor,
	ctx context.Context, acct security.Account, can security.Candidate, auth security.Authentication, err error) (security.Authentication, error) {

	result := AuthenticationResult{
		Candidate: can,
		Auth:      auth,
		Error:     err,
	}
	for _, processor := range processors {
		result = processor.Process(ctx, acct, result)
	}
	return result.Auth, result.Error
}

/******************************
	Common Implementation
 ******************************/

// PersistAccountPostProcessor saves Account. It's implement order.Ordered with highest priority
// Note: post-processors executed in reversed order
type PersistAccountPostProcessor struct {
	store security.AccountStore
}

func NewPersistAccountPostProcessor(store security.AccountStore) *PersistAccountPostProcessor {
	return &PersistAccountPostProcessor{store: store}
}

// Order the processor run last
func (p *PersistAccountPostProcessor) Order() int {
	return postProcessorOrderPersistAccount
}

func (p *PersistAccountPostProcessor) Process(ctx context.Context, acct security.Account, result AuthenticationResult) AuthenticationResult {
	if acct == nil {
		return result
	}

	// regardless decision, account need to be persisted in case of any status changes.
	// Note: we ignore save error since it's too late to do anything
	e := p.store.Save(ctx, acct)
	if e != nil && !errors.Is(e, security.ErrorSubTypeInternalError) {
		logger.WithContext(ctx).Warnf("account status was not persisted due to error: %v", e)
	}
	return result
}

// AccountStatusPostProcessor updates account based on authentication result.
// It could update last login status, failed login status, etc.
type AccountStatusPostProcessor struct {
	store security.AccountStore
}

func NewAccountStatusPostProcessor(store security.AccountStore) *AccountStatusPostProcessor {
	return &AccountStatusPostProcessor{store: store}
}

// Order the processor run first (reversed ordering)
func (p *AccountStatusPostProcessor) Order() int {
	return postProcessorOrderAccountStatus
}

func (p *AccountStatusPostProcessor) Process(ctx context.Context, acct security.Account, result AuthenticationResult) AuthenticationResult {
	updater, ok := acct.(security.AccountUpdater)
	if !ok {
		return result
	}

	switch {
	case result.Error == nil && result.Auth != nil && result.Auth.State() >= security.StateAuthenticated:
		// fully authenticated
		updater.RecordSuccess(time.Now())
		updater.ResetFailedAttempts()
		if history, ok := acct.(security.AccountHistory); ok && history.SerialFailedAttempts() != 0 {
			logger.WithContext(ctx).Warnf("Account [%s] failed to reset", acct.Username())
		}
	case errors.Is(result.Error, errorBadCredentials) && isPasswordAuth(result):
		// Password auth failed with incorrect password
		limit := 5
		if rules, e := p.store.LoadLockingRules(ctx, acct); e == nil && rules != nil && rules.LockoutEnabled() {
			limit = rules.LockoutFailuresLimit()
		}
		updater.RecordFailure(time.Now(), limit)
	default:
	}

	return result
}

// AccountLockingPostProcessor react on failed authentication. Lock account if necessary
type AccountLockingPostProcessor struct {
	store security.AccountStore
}

func NewAccountLockingPostProcessor(store security.AccountStore) *AccountLockingPostProcessor {
	return &AccountLockingPostProcessor{store: store}
}

// Order the processor run between AccountStatusPostProcessor and PersistAccountPostProcessor
func (p *AccountLockingPostProcessor) Order() int {
	return postProcessorOrderAccountLocking
}

func (p *AccountLockingPostProcessor) Process(ctx context.Context, acct security.Account, result AuthenticationResult) AuthenticationResult {
	// skip if
	// 1. account is not updatable
	// 2. not bad credentials
	// 3. not password auth
	updater, ok := acct.(security.AccountUpdater)
	if !ok || !errors.Is(result.Error, errorBadCredentials) || !isPasswordAuth(result) {
		return result
	}

	history, ok := acct.(security.AccountHistory)
	rules, e := p.store.LoadLockingRules(ctx, acct)
	if !ok || e != nil || rules == nil || !rules.LockoutEnabled() {
		return result
	}

	// Note 1: we assume AccountStatusPostProcessor already updated login success/failure records
	// Note 2: we don't count login failure before last lockout time. whether this is necessary is TBD
	// find first login failure within FailureInterval
	now := time.Now()
	count := 0
	for _, t := range history.LoginFailures() {
		if interval := now.Sub(t); interval <= rules.LockoutFailuresInterval() && t.After(history.LockoutTime()) {
			count++
		}
	}

	// lock the account if over the limit
	if count >= rules.LockoutFailuresLimit() {
		updater.LockAccount()
		logger.WithContext(ctx).Infof("Account[%s] Locked", acct.Username())
		// Optional, change error message
		result.Error = security.NewAccountStatusError(MessageLockedDueToBadCredential, result.Error)
	}
	return result
}

// AdditionalDetailsPostProcessor populate additional authentication details if the authentication is granted.
// It's implement order.Ordered
// Note: post-processors executed in reversed order
type AdditionalDetailsPostProcessor struct {}

func NewAdditionalDetailsPostProcessor() *AdditionalDetailsPostProcessor {
	return &AdditionalDetailsPostProcessor{}
}

// Order the processor run last
func (p *AdditionalDetailsPostProcessor) Order() int {
	return postProcessorOrderAdditionalDetails
}

func (p *AdditionalDetailsPostProcessor) Process(_ context.Context, _ security.Account, result AuthenticationResult) AuthenticationResult {
	if result.Error != nil || result.Auth == nil {
		return result
	}
	details, ok := result.Auth.Details().(map[string]interface{})
	if !ok {
		return result
	}

	// auth method
	details[security.DetailsKeyAuthMethod] = security.AuthMethodPassword

	// MFA
	if isMfaVerify(result) {
		details[security.DetailsKeyMFAApplied] = true
	}
	return result
}

/******************************
	Helper
 ******************************/
func isPasswordAuth(result AuthenticationResult) bool {
	_, ok := result.Candidate.(*UsernamePasswordPair)
	return ok
}

func isMfaVerify(result AuthenticationResult) bool {
	_, ok := result.Candidate.(*MFAOtpVerification)
	return ok
}
