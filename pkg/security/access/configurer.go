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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"sort"
)

var (
	FeatureId = security.FeatureId("AC", security.FeatureOrderAccess)
)

//goland:noinspection GoNameStartsWithPackageName
type AccessControlConfigurer struct {

}

func newAccessControlConfigurer() *AccessControlConfigurer {
	return &AccessControlConfigurer{
	}
}

func (acc *AccessControlConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	// Verify
	if err := acc.validate(feature.(*AccessControlFeature), ws); err != nil {
		return err
	}
	f := feature.(*AccessControlFeature)

	// construct decision maker functions
	decisionMakers := make([]DecisionMakerFunc, len(f.acl))
	sort.SliceStable(f.acl, func(i, j int) bool {
		return order.OrderedFirstCompare(f.acl[i], f.acl[j])
	})
	for i, ac := range f.acl {
		if ac.custom != nil {
			decisionMakers[i] = WrapDecisionMakerFunc(ac.matcher, ac.custom)
		} else {
			decisionMakers[i] = MakeDecisionMakerFunc(ac.matcher, ac.control)
		}
	}

	// register middlewares
	mw := NewAccessControlMiddleware(decisionMakers...)
	ac := middleware.NewBuilder("access control").
		Order(security.MWOrderAccessControl).
		Use(mw.ACHandlerFunc())

	ws.Add(ac)
	return nil
}

func (acc *AccessControlConfigurer) validate(f *AccessControlFeature, ws security.WebSecurity) error {
	if len(f.acl) == 0 {
		logger.Infof("access control is not set, default to DenyAll - [%v]", log.Capped(ws, 80))
		f.Request(matcher.AnyRequest()).DenyAll()
	}
	return nil
}
