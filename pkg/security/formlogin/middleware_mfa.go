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

package formlogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	errorutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/error"
	"github.com/gin-gonic/gin"
)

var (

)

type MfaAuthenticationMiddleware struct {
	authenticator  security.Authenticator
	successHandler security.AuthenticationSuccessHandler
	otpParam       string
}

type MfaMWOptionsFunc func(*MfaMWOptions)

type MfaMWOptions struct {
	Authenticator  security.Authenticator
	SuccessHandler security.AuthenticationSuccessHandler
	OtpParam       string
}

func NewMfaAuthenticationMiddleware(optionFuncs ...MfaMWOptionsFunc) *MfaAuthenticationMiddleware {
	options := MfaMWOptions{}
	for _, optFunc := range optionFuncs {
		if optFunc != nil {
			optFunc(&options)
		}
	}
	return &MfaAuthenticationMiddleware{
		authenticator:  options.Authenticator,
		successHandler: options.SuccessHandler,
		otpParam:       options.OtpParam,
	}
}

func (mw *MfaAuthenticationMiddleware) OtpVerifyHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		otp := ctx.PostFormArray(mw.otpParam)
		if len(otp) == 0 {
			otp = []string{""}
		}

		before, err := mw.currentAuth(ctx)
		if err != nil {
			mw.handleError(ctx, err, nil)
			return
		}

		candidate := passwd.MFAOtpVerification{
			CurrentAuth: before,
			OTP:         otp[0],
			DetailsMap:  map[string]interface{}{},
		}

		// authenticate
		auth, err := mw.authenticator.Authenticate(ctx, &candidate)
		if err != nil {
			mw.handleError(ctx, err, &candidate)
			return
		}
		mw.handleSuccess(ctx, before, auth)
	}
}

func (mw *MfaAuthenticationMiddleware) OtpRefreshHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		before, err := mw.currentAuth(ctx)
		if err != nil {
			mw.handleError(ctx, err, nil)
			return
		}
		candidate := passwd.MFAOtpRefresh{
			CurrentAuth: before,
			DetailsMap:  map[string]interface{}{},
		}

		// authenticate
		auth, err := mw.authenticator.Authenticate(ctx, &candidate)
		if err != nil {
			mw.handleError(ctx, err, &candidate)
			return
		}
		mw.handleSuccess(ctx, before, auth)
	}
}

func (mw *MfaAuthenticationMiddleware) currentAuth(ctx *gin.Context) (passwd.UsernamePasswordAuthentication, error) {
	if currentAuth, ok := security.Get(ctx).(passwd.UsernamePasswordAuthentication); !ok || !currentAuth.IsMFAPending() {
		return nil, security.NewAccessDeniedError("MFA is not in progess")
	} else {
		return currentAuth, nil
	}
}

func (mw *MfaAuthenticationMiddleware) handleSuccess(c *gin.Context, before, new security.Authentication) {
	if new != nil {
		c.Set(gin.AuthUserKey, new.Principal())
		c.Set(security.ContextKeySecurity, new)
	}
	mw.successHandler.HandleAuthenticationSuccess(c, c.Request, c.Writer, before, new)
	if c.Writer.Written() {
		c.Abort()
	}
}

func (mw *MfaAuthenticationMiddleware) handleError(c *gin.Context, err error, candidate security.Candidate) {
	if mw.shouldClear(err) {
		security.Clear(c)
	}
	_ = c.Error(err)
	c.Abort()
}

func (mw *MfaAuthenticationMiddleware) shouldClear(err error) bool {
	//nolint:errorlint
	switch coder, ok := err.(errorutils.ErrorCoder); ok {
	case coder.Code() == security.ErrorCodeCredentialsExpired:
		return true
	case coder.Code() == security.ErrorCodeMaxAttemptsReached:
		return true
	}
	return false
}
