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
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/redis"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/utils/order"
    "github.com/pquerna/otp"
    "sort"
    "time"
)

type builderDefaults struct {
	accountStore    security.AccountStore
	passwordEncoder PasswordEncoder
	redisClient     redis.Client
}

// AuthenticatorBuilder implements security.AuthenticatorBuilder
type AuthenticatorBuilder struct {
	feature *PasswordAuthFeature
	defaults *builderDefaults
}

func NewAuthenticatorBuilder(f *PasswordAuthFeature, defaults...*builderDefaults) *AuthenticatorBuilder {
	builder := &AuthenticatorBuilder{
			feature: f,
	}

	if len(defaults) != 0 {
		builder.defaults = defaults[len(defaults) - 1]
	} else {
		builder.defaults = &builderDefaults{}
	}

	return builder
}

func (b *AuthenticatorBuilder) Build(_ context.Context) (security.Authenticator, error) {

	// prepare options
	defaultOpts, err := b.defaultOptions(b.feature)
	if err != nil {
		return nil, err
	}

	mfaOpts, err := b.mfaOptions(b.feature)
	if err != nil {
		return nil, err
	}

	// username passowrd authenticator
	passwdAuth := NewAuthenticator(defaultOpts, mfaOpts)

	// MFA
	if b.feature.mfaEnabled {
		mfaVerify := NewMFAVerifyAuthenticator(defaultOpts, mfaOpts)
		mfaRefresh := NewMFARefreshAuthenticator(defaultOpts, mfaOpts)
		return security.NewAuthenticator(passwdAuth, mfaVerify, mfaRefresh), nil
	}

	return passwdAuth, nil
}

func (b *AuthenticatorBuilder) defaultOptions(f *PasswordAuthFeature) (AuthenticatorOptionsFunc, error) {
	if f.accountStore == nil {
		if b.defaults.accountStore == nil {
			return nil, fmt.Errorf("unable to create password authenticator: account accountStore is not set")
		}
		f.accountStore = b.defaults.accountStore
	}

	if f.passwordEncoder == nil {
		f.passwordEncoder = b.defaults.passwordEncoder
	}

	decisionMakers := b.prepareDecisionMakers(f)
	processors := b.preparePostProcessors(f)

	return func(opts *AuthenticatorOptions) {
		opts.AccountStore = f.accountStore
		if f.passwordEncoder != nil {
			opts.PasswordEncoder = f.passwordEncoder
		}
		opts.Checkers = decisionMakers
		opts.PostProcessors = processors
	}, nil
}

func (b *AuthenticatorBuilder) mfaOptions(f *PasswordAuthFeature) (AuthenticatorOptionsFunc, error) {
	if !f.mfaEnabled {
		return func(*AuthenticatorOptions) {/* noop */}, nil
	}

	if f.otpTTL <= 0 {
		f.otpTTL = 10 * time.Minute
	}

	if f.otpVerifyLimit <= 0 {
		f.otpVerifyLimit = 3
	}

	if f.otpRefreshLimit <= 0 {
		f.otpRefreshLimit = 3
	}

	if f.otpLength <= 3 {
		f.otpLength = 3
	}

	if f.otpSecretSize <= 5 {
		f.otpSecretSize = 5
	}

	otpManager := newTotpManager(func(s *totpManager) {
		s.ttl = f.otpTTL
		s.maxVerifyLimit = f.otpVerifyLimit
		s.maxRefreshLimit = f.otpRefreshLimit
		if b.defaults.redisClient != nil {
			s.store = newRedisOtpStore(b.defaults.redisClient)
		}
		s.factory = newTotpFactory(func(factory *totpFactory) {
			factory.digits = otp.Digits(f.otpLength)
			factory.secretSize = int(f.otpSecretSize)
		})
	})

	decisionMakers := b.prepareDecisionMakers(f)
	processors := b.preparePostProcessors(f)

	return func(opts *AuthenticatorOptions) {
		opts.OTPManager = otpManager
		sort.SliceStable(f.mfaEventListeners, func(i,j int) bool {
			return order.OrderedFirstCompare(f.mfaEventListeners[i], f.mfaEventListeners[j])
		})
		opts.MFAEventListeners = f.mfaEventListeners
		opts.Checkers = decisionMakers
		opts.PostProcessors = processors
	}, nil
}

func (b *AuthenticatorBuilder) prepareDecisionMakers(f *PasswordAuthFeature) []AuthenticationDecisionMaker {
	// maybe customizable via Feature
	acctStatusChecker := NewAccountStatusChecker(f.accountStore)
	passwordChecker := NewPasswordPolicyChecker(f.accountStore)

	return []AuthenticationDecisionMaker{
		PreCredentialsCheck(acctStatusChecker),
		FinalCheck(passwordChecker),
	}
}

func (b *AuthenticatorBuilder) preparePostProcessors(f *PasswordAuthFeature) []PostAuthenticationProcessor {
	// maybe customizable via Feature
	return []PostAuthenticationProcessor{
		NewPersistAccountPostProcessor(f.accountStore),
		NewAdditionalDetailsPostProcessor(),
		NewAccountStatusPostProcessor(f.accountStore),
		NewAccountLockingPostProcessor(f.accountStore),
	}
}









