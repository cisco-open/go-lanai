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
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"reflect"
)

/******************************
	security.Authenticator
******************************/

type Authenticator struct {
	tokenStoreReader oauth2.TokenStoreReader
}

type AuthenticatorOptions func(opt *AuthenticatorOption)

type AuthenticatorOption struct {
	TokenStoreReader oauth2.TokenStoreReader
}

func NewAuthenticator(options ...AuthenticatorOptions) *Authenticator {
	opt := AuthenticatorOption{}
	for _, f := range options {
		if f != nil {
			f(&opt)
		}
	}
	return &Authenticator{
		tokenStoreReader:      opt.TokenStoreReader,
	}
}

func (a *Authenticator) Authenticate(ctx context.Context, candidate security.Candidate) (security.Authentication, error) {
	can, ok := candidate.(*BearerToken)
	if !ok {
		return nil, nil
	}

	// TODO add remote check_token endpoint support
	auth, e := a.tokenStoreReader.ReadAuthentication(ctx, can.Token, oauth2.TokenHintAccessToken)
	if e != nil {
		return nil, e
	}

	// perform some checks
	switch {
	case auth.State() < security.StateAuthenticated:
		return nil, oauth2.NewInvalidAccessTokenError("token is not associated with an authenticated session")
	case auth.OAuth2Request().ClientId() == "":
		return nil, oauth2.NewInvalidAccessTokenError("token is not issued to a valid client")
	case auth.UserAuthentication() != nil && reflect.ValueOf(auth.UserAuthentication().Principal()).IsZero():
		return nil, oauth2.NewInvalidAccessTokenError("token is not authorized by a valid user")
	}

	return auth, nil
}