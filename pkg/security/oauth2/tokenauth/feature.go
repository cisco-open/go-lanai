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
)

//goland:noinspection GoNameStartsWithPackageName
type TokenAuthFeature struct {
	errorHandler    *OAuth2ErrorHandler
	postBodyEnabled bool
}

func (f *TokenAuthFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

// Configure Standard security.Feature entrypoint
// use (*access.AccessControl).AllowIf(ScopesApproved(...)) for scope based access decision maker
func Configure(ws security.WebSecurity) *TokenAuthFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*TokenAuthFeature)
	}
	panic(fmt.Errorf("unable to configure oauth2 authserver: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// New Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
// use (*access.AccessControl).AllowIf(ScopesApproved(...)) for scope based access decision maker
func New() *TokenAuthFeature {
	return &TokenAuthFeature{}
}

/** Setters **/

func (f *TokenAuthFeature) ErrorHandler(errorHandler *OAuth2ErrorHandler) *TokenAuthFeature {
	f.errorHandler = errorHandler
	return f
}

func (f *TokenAuthFeature) EnablePostBody() *TokenAuthFeature {
	f.postBodyEnabled = true
	return f
}
