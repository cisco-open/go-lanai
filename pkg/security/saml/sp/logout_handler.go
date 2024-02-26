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

package sp

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
	"net/http"
)

var ErrSamlSloRequired = security.NewAuthenticationError("SAML SLO required")

type SingleLogoutHandler struct{}

func NewSingleLogoutHandler() *SingleLogoutHandler {
	return &SingleLogoutHandler{}
}

// ShouldLogout is a logout.ConditionalLogoutHandler method that interrupt logout process by returning authentication error,
// which would trigger authentication entry point and initiate SLO
func (h *SingleLogoutHandler) ShouldLogout(ctx context.Context, _ *http.Request, _ http.ResponseWriter, auth security.Authentication) error {
	if !h.requiresSamlSLO(ctx, auth) {
		return nil
	}
	return ErrSamlSloRequired
}

func (h *SingleLogoutHandler) HandleLogout(ctx context.Context, _ *http.Request, _ http.ResponseWriter, auth security.Authentication) error {
	if !h.wasSLOFailed(ctx, auth) {
		return nil
	}
	return security.NewAuthenticationWarningError("cisco.saml.logout.failed")
}

func (h *SingleLogoutHandler) samlDetails(_ context.Context, auth security.Authentication) (map[string]interface{}, bool) {
	switch v := auth.(type) {
	case *samlAssertionAuthentication:
		return v.DetailsMap, true
	default:
		m, _ := auth.Details().(map[string]interface{})
		return m, false
	}
}

func (h *SingleLogoutHandler) requiresSamlSLO(ctx context.Context, auth security.Authentication) bool {
	var isSaml, sloCompleted bool
	var details map[string]interface{}
	// check if it's saml
	details, isSaml = h.samlDetails(ctx, auth)

	// check if SLO already completed
	state, ok := details[kDetailsSLOState].(SLOState)
	sloCompleted = ok && state.Is(SLOCompleted)

	return isSaml && !sloCompleted
}

func (h *SingleLogoutHandler) wasSLOFailed(ctx context.Context, auth security.Authentication) bool {
	var isSaml, sloFailed bool
	var details map[string]interface{}
	details, isSaml = h.samlDetails(ctx, auth)

	// check if SLO already completed
	state, ok := details[kDetailsSLOState].(SLOState)
	sloFailed = ok && state.Is(SLOFailed)

	return isSaml && sloFailed
}
