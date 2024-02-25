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

package csrf

import (
    "context"
    "encoding/gob"
    "errors"
    "github.com/cisco-open/go-lanai/pkg/security/session"
    "github.com/google/uuid"
)

const SessionKeyCsrfToken = "CsrfToken"

// Token CSRF token with value and other useful metadata
/**
 The header name and parameter name are part of the token in case some components down the line needs them.
 For example, if the token is used as a hidden variable in a form, the parameter name would be needed.
 */
type Token struct {
	Value string

	// the HTTP parameter that the CSRF token can be placed on request
	ParameterName string

	// the HTTP header that the CSRF can be placed on requests instead of the parameter.
	HeaderName string
}

// TokenStore
/**
	The store is responsible for reading the CSRF token associated to the request.
	How the CSRF token is associated to the request is the implementation's discretion.

	The store is responsible for writing to the response header if necessary
	for example, if the store implementation is based on cookies, then the save method
	would write (save) the token as a cookie header.
 */
type TokenStore interface {
	Generate(c context.Context, parameterName string, headerName string) *Token

	SaveToken(c context.Context, token *Token) error

	LoadToken(c context.Context) (*Token, error)
}

type SessionBackedStore struct {
}

func newSessionBackedStore() *SessionBackedStore{
	gob.Register((*Token)(nil))
	return &SessionBackedStore{}
}

func (store *SessionBackedStore) Generate(c context.Context, parameterName string, headerName string) *Token {
	t := &Token{
		Value: uuid.New().String(),
		ParameterName: parameterName,
		HeaderName: headerName,
	}
	return t
}

func (store *SessionBackedStore) SaveToken(c context.Context, token *Token) error {
	s := session.Get(c)

	if s == nil {
		return errors.New("can't save csrf token to session, because the request has no session")
	}

	s.Set(SessionKeyCsrfToken, token)
	return s.Save()
}

func (store *SessionBackedStore) LoadToken(c context.Context) (*Token, error) {
	s := session.Get(c)

	if s == nil {
		return nil, errors.New("can't load csrf token from session, because the request has no session")
	}

	attr := s.Get(SessionKeyCsrfToken)

	if attr == nil {
		return nil, nil
	}

	if token, ok := attr.(*Token); !ok {
		return nil, errors.New("csrf token in session can't be asserted to be the CSRF token type")
	} else {
		return token, nil
	}
}