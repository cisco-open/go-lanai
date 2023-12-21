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
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"time"
)

var b32NoPadding = base32.StdEncoding.WithPadding(base32.NoPadding)

type TOTP struct {
	Passcode string
	Secret string
	TTL time.Duration
	Expire time.Time
}

type TOTPFactory interface {
	Generate(ttl time.Duration) (totp TOTP, err error)
	Refresh(secret string, ttl time.Duration) (totp TOTP, err error)
	Validate(totp TOTP) (valid bool, err error)
}

type totpFactory struct {
	skew       uint
	digits     otp.Digits
	alg        otp.Algorithm
	secretSize int
}

type totpFactoryOptions func(*totpFactory)

func newTotpFactory(options...totpFactoryOptions) *totpFactory {
	factory := &totpFactory{
		skew: 0,
		digits: 6,
		alg: otp.AlgorithmSHA1,
		secretSize: 20,
	}

	for _,opt := range options {
		opt(factory)
	}
	return factory
}

func (f *totpFactory) Generate(ttl time.Duration) (ret TOTP, err error) {
	secret, err := f.generateSecret()
	if err != nil {
		return
	}

	return f.Refresh(secret, ttl)
}

func (f *totpFactory) Refresh(secret string, ttl time.Duration) (ret TOTP, err error) {
	if ttl < time.Second {
		return ret, fmt.Errorf("ttl should be greater or equals to 1 seconds")
	}

	now := time.Now()
	ttl = ttl.Round(time.Second)
	passcode, err := totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{
		Period:    uint(ttl.Seconds()),
		Skew:      f.skew,
		Digits:    f.digits,
		Algorithm: f.alg,
	})
	if err != nil {
		return
	}

	ret = TOTP{
		Passcode: passcode,
		Secret: secret,
		TTL: ttl,
		Expire: now.Add(ttl),
	}
	return
}

func (f *totpFactory) Validate(value TOTP) (valid bool, err error) {
	if value.TTL < time.Second {
		return false, fmt.Errorf("ttl should be greater or equals to 1 seconds")
	}

	return totp.ValidateCustom(value.Passcode, value.Secret, time.Now(), totp.ValidateOpts{
		Period:    uint(value.TTL.Round(time.Second).Seconds()),
		Skew:      f.skew,
		Digits:    f.digits,
		Algorithm: f.alg,
	})
}

func (f *totpFactory) generateSecret() (string, error) {
	secret := make([]byte, f.secretSize)

	_, err := rand.Reader.Read(secret)
	if err != nil {
		return "", err
	}
	return b32NoPadding.EncodeToString(secret), nil
}
