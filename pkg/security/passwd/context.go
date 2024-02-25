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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"encoding/gob"
	"time"
)

const (
	SpecialPermissionMFAPending = "MFAPending"
	SpecialPermissionOtpId      = "OtpId"
)

const (
	MessageUserNotFound              = "Mismatched Username and Password"
	MessageBadCredential             = "Mismatched Username and Password"
	MessageOtpNotAvailable           = "MFA required but temprorily unavailable"
	MessageAccountStatus             = "Inactive Account"
	MessageInvalidPasscode           = "Bad Verification Code"
	MessagePasscodeExpired           = "Verification Code Expired"
	MessageCannotRefresh             = "Unable to Refresh"
	MessageMaxAttemptsReached        = "No More Verification Attempts Allowed"
	MessageMaxRefreshAttemptsReached = "No More Resend Attempts Allowed"
	MessageInvalidAccountStatus      = "Issue with current account status"
	MessageAccountDisabled           = "Account Disabled"
	MessageAccountLocked             = "Account Locked"
	MessagePasswordLoginNotAllowed   = "Password Login not Allowed"
	MessageLockedDueToBadCredential  = "Mismatched Username and Password. Account locked due to too many failed attempts"
	MessagePasswordExpired           = "User credentials have expired"
)

// For error translation
var (
	errorBadCredentials     = security.NewBadCredentialsError("bad creds")
	errorCredentialsExpired = security.NewCredentialsExpiredError("cred exp")
	errorMaxAttemptsReached = security.NewMaxAttemptsReachedError("max attempts")
	//errorAccountStatus      = security.NewAccountStatusError("acct status")
)

/******************************
	Serialization
******************************/

func GobRegister() {
	gob.Register((*usernamePasswordAuthentication)(nil))
	gob.Register((*timeBasedOtp)(nil))
	gob.Register(TOTP{})
	gob.Register(time.Time{})
	gob.Register(time.Duration(0))
}

/************************
	security.Candidate
************************/

type MFAMode int

const (
	MFAModeSkip MFAMode = iota
	MFAModeOptional
	MFAModeMust
)

// UsernamePasswordPair is the supported security.Candidate of this authenticator
type UsernamePasswordPair struct {
	Username   string
	Password   string
	DetailsMap map[string]interface{}
	EnforceMFA MFAMode
}

// Principal implements security.Candidate
func (upp *UsernamePasswordPair) Principal() interface{} {
	return upp.Username
}

// Credentials implements security.Candidate
func (upp *UsernamePasswordPair) Credentials() interface{} {
	return upp.Password
}

// Details implements security.Candidate
func (upp *UsernamePasswordPair) Details() interface{} {
	return upp.DetailsMap
}

// MFAOtpVerification is the supported security.Candidate for MFA authentication
type MFAOtpVerification struct {
	CurrentAuth UsernamePasswordAuthentication
	OTP         string
	DetailsMap  map[string]interface{}
}

// Principal implements security.Candidate
func (uop *MFAOtpVerification) Principal() interface{} {
	return uop.CurrentAuth.Principal()
}

// Credentials implements security.Candidate
func (uop *MFAOtpVerification) Credentials() interface{} {
	return uop.OTP
}

// Details implements security.Candidate
func (uop *MFAOtpVerification) Details() interface{} {
	return uop.DetailsMap
}

// MFAOtpRefresh is the supported security.Candidate for MFA OTP refresh
type MFAOtpRefresh struct {
	CurrentAuth UsernamePasswordAuthentication
	DetailsMap  map[string]interface{}
}

// Principal implements security.Candidate
func (uop *MFAOtpRefresh) Principal() interface{} {
	return uop.CurrentAuth.Principal()
}

// Credentials implements security.Candidate
func (uop *MFAOtpRefresh) Credentials() interface{} {
	return uop.CurrentAuth.OTPIdentifier()
}

// Details implements security.Candidate
func (uop *MFAOtpRefresh) Details() interface{} {
	return uop.DetailsMap
}

/******************************
	security.Authentication
******************************/

// UsernamePasswordAuthentication implements security.Authentication
type UsernamePasswordAuthentication interface {
	security.Authentication
	Username() string
	IsMFAPending() bool
	OTPIdentifier() string
}

// TODO: do we want the details here to also implement the ctx_details interfaces?
// usernamePasswordAuthentication
// Note: all fields should not be used directly. It's exported only because gob only deal with exported field
type usernamePasswordAuthentication struct {
	Acct       security.Account
	Perms      map[string]interface{}
	DetailsMap map[string]interface{}
}

func (auth *usernamePasswordAuthentication) Principal() interface{} {
	return auth.Acct
}

func (auth *usernamePasswordAuthentication) Permissions() security.Permissions {
	return auth.Perms
}

func (auth *usernamePasswordAuthentication) State() security.AuthenticationState {
	switch {
	case auth.IsMFAPending():
		return security.StatePrincipalKnown
	default:
		return security.StateAuthenticated
	}
}

func (auth *usernamePasswordAuthentication) Details() interface{} {
	return auth.DetailsMap
}

func (auth *usernamePasswordAuthentication) Username() string {
	return auth.Acct.Username()
}

func (auth *usernamePasswordAuthentication) IsMFAPending() bool {
	_, ok := auth.Permissions()[SpecialPermissionOtpId].(string)
	return ok
}

func (auth *usernamePasswordAuthentication) OTPIdentifier() string {
	v, ok := auth.Permissions()[SpecialPermissionOtpId].(string)
	if ok {
		return v
	}
	return ""
}

func IsSamePrincipal(username string, currentAuth security.Authentication) bool {
	if currentAuth == nil || currentAuth.State() < security.StatePrincipalKnown {
		return false
	}

	if account, ok := currentAuth.Principal().(security.Account); ok && username == account.Username() {
		return true
	} else if principal, ok := currentAuth.Principal().(string); ok && username == principal {
		return true
	}
	return false
}
