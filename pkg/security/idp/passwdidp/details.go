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

package passwdidp

import "github.com/cisco-open/go-lanai/pkg/security/idp"

type PasswdIdpDetails struct {
	Domain           string
}


type PasswdIdpOptions func(opt *PasswdIdpDetails)

// PasswdIdentityProvider implements idp.IdentityProvider and idp.AuthenticationFlowAware
type PasswdIdentityProvider struct {
	PasswdIdpDetails
}

func NewIdentityProvider(opts ...PasswdIdpOptions) *PasswdIdentityProvider {
	opt := PasswdIdpDetails{}
	for _, f := range opts {
		f(&opt)
	}
	return &PasswdIdentityProvider{
		PasswdIdpDetails: opt,
	}
}

func (s PasswdIdentityProvider) AuthenticationFlow() idp.AuthenticationFlow {
	return idp.InternalIdpForm
}

func (s PasswdIdentityProvider) Domain() string {
	return s.PasswdIdpDetails.Domain
}
