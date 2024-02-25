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
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/passwd"
	"net/http"
)

type MfaAwareAuthenticationEntryPoint struct {
	delegate security.AuthenticationEntryPoint
	mfaPendingDelegate security.AuthenticationEntryPoint
}

func (h *MfaAwareAuthenticationEntryPoint) Commence(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	auth,ok := security.Get(c).(passwd.UsernamePasswordAuthentication)
	if ok && auth.IsMFAPending() {
		h.mfaPendingDelegate.Commence(c, r, rw, err)
	} else {
		h.delegate.Commence(c, r, rw, err)
	}
}

type MfaAwareSuccessHandler struct {
	delegate security.AuthenticationSuccessHandler
	mfaPendingDelegate security.AuthenticationSuccessHandler
}

func (h *MfaAwareSuccessHandler) HandleAuthenticationSuccess(
	c context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	userAuth,ok := to.(passwd.UsernamePasswordAuthentication)
	if ok && userAuth.IsMFAPending() {
		h.mfaPendingDelegate.HandleAuthenticationSuccess(c, r, rw, from, to)
	} else {
		h.delegate.HandleAuthenticationSuccess(c, r, rw, from, to)
	}
}

type MfaAwareAuthenticationErrorHandler struct {
	delegate security.AuthenticationErrorHandler
	mfaPendingDelegate security.AuthenticationErrorHandler
}

func (h *MfaAwareAuthenticationErrorHandler) HandleAuthenticationError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	auth,ok := security.Get(c).(passwd.UsernamePasswordAuthentication)
	if ok && auth.IsMFAPending() {
		h.mfaPendingDelegate.HandleAuthenticationError(c, r, rw, err)
	} else {
		h.delegate.HandleAuthenticationError(c, r, rw, err)
	}
}
