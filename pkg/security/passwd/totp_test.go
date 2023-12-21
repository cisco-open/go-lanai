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
	"testing"
	"time"
)


var otpFactory TOTPFactory = newTotpFactory(func(f *totpFactory) {
	f.digits = 6
	f.secretSize = 20
})

func TestOtpGeneration(t *testing.T) {
	now := time.Now()
	expectedExpire := now.Add(time.Second * 10);
	totp, err := otpFactory.Generate(time.Second * 10)
	switch {
	case err != nil:
		t.Errorf("Generate should not return error")

	case len(totp.Secret) != 32:
		t.Errorf("Generate should return 32 charaters of secret")

	case len(totp.Passcode) != 6:
		t.Errorf("Generate should returns 6 digits passcode")

	case totp.TTL != time.Second * 10:
		t.Errorf("Refresh should returns correct TTL round to seconds")

	case totp.Expire.Before(expectedExpire):
		t.Errorf("Generate should not return expire time before expected validation period")

	case totp.Expire.After(expectedExpire.Add(time.Second)):
		t.Errorf("Generate should not return expire time after 1s past expected validation period")
	}
}

func TestOtpRefresh(t *testing.T) {
	now := time.Now()
	original, _ := otpFactory.Generate(time.Second * 10)
	totp, err := otpFactory.Refresh(original.Secret, time.Second * 5)
	expectedExpire := now.Add(time.Second * 5);
	switch {
	case err != nil:
		t.Errorf("Refresh should not return error")

	case totp.Secret != original.Secret:
		t.Errorf("Refresh should return same secret as original")

	case len(totp.Passcode) != 6:
		t.Errorf("Refresh should returns 6 digits passcode")

	case totp.Passcode == original.Passcode:
		t.Errorf("Refresh should returns different passcode")

	case totp.TTL != time.Second * 5:
		t.Errorf("Refresh should returns correct TTL round to seconds")

	case totp.Expire.Before(expectedExpire):
		t.Errorf("Generate should not return expire time before expected validation period")

	case totp.Expire.After(expectedExpire.Add(time.Second)):
		t.Errorf("Generate should not return expire time after 1s past expected validation period")
	}
}

func TestValidateGeneratedCode(t *testing.T) {
	totp, _ := otpFactory.Generate(time.Second * 10)
	valid, err := otpFactory.Validate(totp)

	switch {
	case err != nil:
		t.Errorf("Verify should not return error")

	case !valid:
		t.Errorf("Verify should return valid result")
	}

	valid, err = otpFactory.Validate(TOTP{Secret: totp.Secret, Passcode: "000000", TTL: time.Second * 10, Expire: totp.Expire})
	switch {
	case err != nil:
		t.Errorf("Verify should not return error")

	case valid:
		t.Errorf("Verify should return invalid result for wrong passcode")
	}

	valid, err = otpFactory.Validate(TOTP{Secret: "abcdefg", Passcode: totp.Passcode, TTL: time.Second * 10, Expire: totp.Expire})
	switch {
	case err != nil:
		t.Errorf("Verify should not return error")

	case valid:
		t.Errorf("Verify should return invalid result for wrong secret")
	}

	totp, _ = otpFactory.Generate(time.Second * 1)
	time.Sleep(time.Millisecond * 1050)
	valid, err = otpFactory.Validate(totp)
	switch {
	case err != nil:
		t.Errorf("Verify should not return error")

	case valid:
		t.Errorf("Verify should return invalid result for expired OTP")
	}
}

func TestValidateRefreshedCode(t *testing.T) {
	original, _ := otpFactory.Generate(time.Second * 10)
	totp, err := otpFactory.Refresh(original.Secret, time.Second * 5)
	valid, err := otpFactory.Validate(totp)

	switch {
	case err != nil:
		t.Errorf("Verify should not return error")

	case !valid:
		t.Errorf("Verify should return valid result")
	}

	totp, _ = otpFactory.Refresh(original.Secret, time.Second * 1)
	time.Sleep(time.Millisecond * 1050)
	valid, err = otpFactory.Validate(totp)
	switch {
	case err != nil:
		t.Errorf("Verify should not return error")

	case valid:
		t.Errorf("Verify should return invalid result for expired OTP")
	}
}