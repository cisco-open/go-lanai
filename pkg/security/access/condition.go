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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"fmt"
)

// ControlCondition extends web.RequestMatcher, and matcher.ChainableMatcher
// it is used together with web.RoutedMapping's "Condition" for a convienent config of securities
// only matcher.ChainableMatcher's .MatchesWithContext (context.Context, interface{}) (bool, error) is used
// Matches(interface{}) (bool, error) should return regular as if the context is empty
//
// In addition, implementation should also return AccessDeniedError when condition didn't match.
// web.Registrar will propagate this error along the handler chain until it's handled by errorhandling middleware
type ControlCondition matcher.ChainableMatcher

/**************************
	Common Impl.
***************************/

// ConditionWithControlFunc is a common ControlCondition implementation backed by ControlFunc
type ConditionWithControlFunc struct {
	Description string
	ControlFunc ControlFunc
}

func (m *ConditionWithControlFunc) Matches(i interface{}) (bool, error) {
	return m.MatchesWithContext(context.Background(), i)
}

func (m *ConditionWithControlFunc) MatchesWithContext(c context.Context, _ interface{}) (bool, error) {
	auth := security.Get(c)
	return m.ControlFunc(auth)
}

func (m *ConditionWithControlFunc) Or(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.Or(m, matchers...)
}

func (m *ConditionWithControlFunc) And(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.And(m, matchers...)
}

func (m ConditionWithControlFunc) String() string {
	switch {
	case len(m.Description) != 0:
		return m.Description
	default:
		return "access.ControlCondition"
	}
}

/**************************
	Constructors
***************************/

// RequirePermissions returns ControlCondition using HasPermissionsWithExpr
// e.g. RequirePermissions("P1 && P2 && !(P3 || P4)"), means security.Permissions contains both P1 and P2 but not contains neither P3 nor P4
// see HasPermissionsWithExpr for expression syntax
func RequirePermissions(expr string) ControlCondition {
	return &ConditionWithControlFunc{
		Description: fmt.Sprintf("user's permissions match [%s]", expr),
		ControlFunc: HasPermissionsWithExpr(expr),
	}
}

/**************************
	Helpers
***************************/

