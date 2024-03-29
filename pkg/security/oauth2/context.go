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

package oauth2

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
)

/******************************
	security.Authentication
******************************/

// Authentication extends security.Authentication
type Authentication interface {
	security.Authentication
	UserAuthentication() security.Authentication
	OAuth2Request() OAuth2Request
	AccessToken() AccessToken
}

type AuthenticationOptions func(opt *AuthOption)

type AuthOption struct {
	Request  OAuth2Request
	UserAuth security.Authentication
	Token    AccessToken
	Details  interface{}
}

// authentication
type authentication struct {
	Request   OAuth2Request                `json:"request"`
	UserAuth  security.Authentication      `json:"userAuth"`
	AuthState security.AuthenticationState `json:"state"`
	token     AccessToken
	details   interface{}
}

func NewAuthentication(opts ...AuthenticationOptions) Authentication {
	config := AuthOption{}
	for _, opt := range opts {
		opt(&config)
	}
	return &authentication{
		Request:   config.Request,
		UserAuth:  config.UserAuth,
		AuthState: calculateState(config.Request, config.UserAuth),
		token:     config.Token,
		details:   config.Details,
	}
}

func (a *authentication) Principal() interface{} {
	if a.UserAuth == nil {
		return a.Request.ClientId()
	}
	return a.UserAuth.Principal()
}

func (a *authentication) Permissions() security.Permissions {
	if a.UserAuth == nil {
		return map[string]interface{}{}
	}
	return a.UserAuth.Permissions()
}

func (a *authentication) State() security.AuthenticationState {
	return a.AuthState
}

func (a *authentication) Details() interface{} {
	return a.details
}

func (a *authentication) UserAuthentication() security.Authentication {
	return a.UserAuth
}

func (a *authentication) OAuth2Request() OAuth2Request {
	return a.Request
}

func (a *authentication) AccessToken() AccessToken {
	return a.token
}

func calculateState(req OAuth2Request, userAuth security.Authentication) security.AuthenticationState {
	if req != nil && req.Approved() {
		if userAuth != nil {
			return userAuth.State()
		}
		return security.StateAuthenticated
	} else if userAuth != nil {
		return security.StatePrincipalKnown
	}
	return security.StateAnonymous
}

/******************************
	UserAuthentication
******************************/

type UserAuthentication interface {
	security.Authentication
	Subject() string
	DetailsMap() map[string]interface{}
}

type UserAuthOptions func(opt *UserAuthOption)

type UserAuthOption struct {
	Principal   string
	Permissions map[string]interface{}
	State       security.AuthenticationState
	Details     map[string]interface{}
}

// userAuthentication implements security.Authentication and UserAuthentication.
// it represents basic information that could be typically extracted from JWT claims
// userAuthentication is also used for serializing/deserializing
type userAuthentication struct {
	SubjectVal    string                       `json:"principal"`
	PermissionVal map[string]interface{}       `json:"permissions"`
	StateVal      security.AuthenticationState `json:"state"`
	DetailsVal    map[string]interface{}       `json:"details"`
}

func NewUserAuthentication(opts ...UserAuthOptions) *userAuthentication {
	opt := UserAuthOption{
		Permissions: map[string]interface{}{},
		Details:     map[string]interface{}{},
	}
	for _, f := range opts {
		f(&opt)
	}
	return &userAuthentication{
		SubjectVal:    opt.Principal,
		PermissionVal: opt.Permissions,
		StateVal:      opt.State,
		DetailsVal:    opt.Details,
	}
}

func (a *userAuthentication) Principal() interface{} {
	return a.SubjectVal
}

func (a *userAuthentication) Permissions() security.Permissions {
	return a.PermissionVal
}

func (a *userAuthentication) State() security.AuthenticationState {
	return a.StateVal
}

func (a *userAuthentication) Details() interface{} {
	return a.DetailsVal
}

func (a *userAuthentication) Subject() string {
	return a.SubjectVal
}

func (a *userAuthentication) DetailsMap() map[string]interface{} {
	return a.DetailsVal
}

/*********************
	Timeout Support
*********************/

type TimeoutApplier interface {
	ApplyTimeout(ctx context.Context, sessionId string) (valid bool, err error)
}
