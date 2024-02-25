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

package tokenauth

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/access"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2"
)

/************************
	Access Control
************************/

func ScopesApproved(scopes...string) access.ControlFunc {
	if len(scopes) == 0 {
		return func(_ security.Authentication) (bool, error) {
			return true, nil
		}
	}

	return func(auth security.Authentication) (decision bool, reason error) {
		err := security.NewAccessDeniedError("required scope was not approved by user")
		switch oauth := auth.(type) {
		case oauth2.Authentication:
			if oauth.OAuth2Request() == nil || !oauth.OAuth2Request().Approved() {
				return false, err
			}

			approved := oauth.OAuth2Request().Scopes()
			if approved == nil || !approved.HasAll(scopes...) {
				return false, err
			}
		default:
			return false, err
		}
		return true, nil
	}
}

/******************************
	Access Control Conditions
*******************************/

// RequireScopes returns ControlCondition using ScopesApproved
func RequireScopes(scopes ...string) access.ControlCondition {
	return &access.ConditionWithControlFunc{
		Description:   fmt.Sprintf("client has scopes [%s] approved", scopes),
		ControlFunc:   ScopesApproved(scopes...),
	}
}