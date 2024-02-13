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

package sectest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/gin-gonic/gin"
)

type MockAuthenticationMiddleware struct {
	MWMocker             MWMocker
	// deprecated, use MWMocker interface or MWMockFunc.
	// Recommended to use WithMockedMiddleware test options
	MockedAuthentication security.Authentication
}

// NewMockAuthenticationMiddleware
// Deprecated, directly set MWMocker field with MWMocker interface or MWMockFunc, Recommended to use WithMockedMiddleware test options
func NewMockAuthenticationMiddleware(authentication security.Authentication) *MockAuthenticationMiddleware {
	return &MockAuthenticationMiddleware{
		MockedAuthentication: authentication,
		MWMocker: MWMockFunc(func(MWMockContext) security.Authentication {
			return authentication
		}),
	}
}

func (m *MockAuthenticationMiddleware) AuthenticationHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var auth security.Authentication
		if m.MWMocker != nil {
			auth = m.MWMocker.Mock(MWMockContext{
				Request: ctx.Request,
			})
		}
		if auth == nil {
			auth = m.MockedAuthentication
		}
		security.MustSet(ctx, auth)
	}
}

type MockUserAuthOptions func(opt *MockUserAuthOption)

type MockUserAuthOption struct {
	Principal   string
	Permissions map[string]interface{}
	State       security.AuthenticationState
	Details     interface{}
}

type mockUserAuthentication struct {
	Subject       string
	PermissionMap map[string]interface{}
	StateValue    security.AuthenticationState
	details       interface{}
}

func NewMockedUserAuthentication(opts ...MockUserAuthOptions) *mockUserAuthentication {
	opt := MockUserAuthOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &mockUserAuthentication{
		Subject:       opt.Principal,
		PermissionMap: opt.Permissions,
		StateValue:    opt.State,
		details:       opt.Details,
	}
}

func (a *mockUserAuthentication) Principal() interface{} {
	return a.Subject
}

func (a *mockUserAuthentication) Permissions() security.Permissions {
	return a.PermissionMap
}

func (a *mockUserAuthentication) State() security.AuthenticationState {
	return a.StateValue
}

func (a *mockUserAuthentication) Details() interface{} {
	return a.details
}
