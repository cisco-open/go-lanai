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

package access

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/web"
	"net/http"
)

// DecisionMakerFunc determine if current user can access to given http.Request
// if the given request is not handled by this function, return false, nil
// if the given request is handled and the access is granted, return true, nil
// otherwise, return true, security.ErrorTypeCodeAccessControl error
type DecisionMakerFunc func(context.Context, *http.Request) (handled bool, decision error)

// AcrMatcher short for Access Control RequestDetails Matcher, accepts *http.Request or http.Request
type AcrMatcher web.RequestMatcher

// ControlFunc make access control decision based on security.Authentication
// "decision" indicate whether the access is grated
// "reason" is optional and is used when access is denied. if not specified, security.NewAccessDeniedError will be used
type ControlFunc func(security.Authentication) (decision bool, reason error)

func MakeDecisionMakerFunc(matcher AcrMatcher, cf ControlFunc) DecisionMakerFunc {
	return func(ctx context.Context, r *http.Request) (bool, error) {
		matches, err := matcher.MatchesWithContext(ctx, r)
		if !matches || err != nil {
			return false, err
		}

		auth := security.Get(ctx)
		granted, reason := cf(auth)
		switch {
		case granted:
			return true, nil
		case reason != nil:
			return true, reason
		default:
			return true, security.NewAccessDeniedError("access denied")
		}
	}
}

func WrapDecisionMakerFunc(matcher AcrMatcher, dmf DecisionMakerFunc) DecisionMakerFunc {
	return func(ctx context.Context, r *http.Request) (bool, error) {
		matches, err := matcher.MatchesWithContext(ctx, r)
		if !matches || err != nil {
			return false, err
		}
		return dmf(ctx, r)
	}
}
