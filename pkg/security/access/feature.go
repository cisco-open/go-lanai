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
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
)

//goland:noinspection GoNameStartsWithPackageName
type AccessControl struct {
	owner   *AccessControlFeature
	order   int
	matcher AcrMatcher
	control ControlFunc
	custom  DecisionMakerFunc
}

// Order implements order.Ordered
func (ac *AccessControl) Order() int {
	return ac.order
}

func (ac *AccessControl) WithOrder(order int) *AccessControl {
	ac.order = order
	return ac
}

func (ac *AccessControl) PermitAll() *AccessControlFeature {
	ac.control = PermitAll
	return ac.owner
}

func (ac *AccessControl) DenyAll() *AccessControlFeature {
	ac.control = DenyAll
	return ac.owner
}

func (ac *AccessControl) Authenticated() *AccessControlFeature {
	ac.control = Authenticated
	return ac.owner
}

func (ac *AccessControl) HasPermissions(permissions ...string) *AccessControlFeature {
	ac.control = HasPermissions(permissions...)
	return ac.owner
}

func (ac *AccessControl) AllowIf(cf ControlFunc) *AccessControlFeature {
	ac.control = cf
	return ac.owner
}

// CustomDecisionMaker override ControlFunc. Order and AcrMatcher are still applied
func (ac *AccessControl) CustomDecisionMaker(dmf DecisionMakerFunc) *AccessControlFeature {
	ac.custom = dmf
	return ac.owner
}

//goland:noinspection GoNameStartsWithPackageName
type AccessControlFeature struct {
	acl []*AccessControl
}

// Identifier implements security.Feature
func (f *AccessControlFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

// Request configure access control of requests matching given AcrMatcher
func (f *AccessControlFeature) Request(matcher AcrMatcher) *AccessControl {
	ac := &AccessControl{
		owner:   f,
		matcher: matcher,
	}
	f.acl = append(f.acl, ac)
	return ac
}

func Configure(ws security.WebSecurity) *AccessControlFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*AccessControlFeature)
	}
	panic(fmt.Errorf("unable to configure access control: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// New Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *AccessControlFeature {
	return &AccessControlFeature{}
}

/**************************
	Common ControlFunc
***************************/

func PermitAll(_ security.Authentication) (bool, error) {
	return true, nil
}

func DenyAll(_ security.Authentication) (bool, error) {
	return false, nil
}

func Authenticated(auth security.Authentication) (bool, error) {
	if auth.State() >= security.StateAuthenticated {
		return true, nil
	}
	return false, security.NewInsufficientAuthError("not fully authenticated")
}

// Note: More ControlFunc can be found in permissions.go
