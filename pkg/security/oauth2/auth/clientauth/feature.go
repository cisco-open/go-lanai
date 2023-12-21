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

package clientauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"fmt"
)

// We currently don't have any stuff to configure
//goland:noinspection GoNameStartsWithPackageName
type ClientAuthFeature struct {
	clientStore         oauth2.OAuth2ClientStore
	clientSecretEncoder passwd.PasswordEncoder
	errorHandler        *auth.OAuth2ErrorHandler
	allowForm           bool
}

// Standard security.Feature entrypoint
func (f *ClientAuthFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func Configure(ws security.WebSecurity) *ClientAuthFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*ClientAuthFeature)
	}
	panic(fmt.Errorf("unable to configure oauth2 authserver: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *ClientAuthFeature {
	return &ClientAuthFeature{
	}
}

/** Setters **/
func (f *ClientAuthFeature) ClientStore(clientStore oauth2.OAuth2ClientStore) *ClientAuthFeature {
	f.clientStore = clientStore
	return f
}

func (f *ClientAuthFeature) ClientSecretEncoder(clientSecretEncoder passwd.PasswordEncoder) *ClientAuthFeature {
	f.clientSecretEncoder = clientSecretEncoder
	return f
}

func (f *ClientAuthFeature) ErrorHandler(errorHandler *auth.OAuth2ErrorHandler) *ClientAuthFeature {
	f.errorHandler = errorHandler
	return f
}

// AllowForm with "true" also implicitly enables Public Client (client that with empty secret)
func (f *ClientAuthFeature) AllowForm(allowForm bool) *ClientAuthFeature {
	f.allowForm = allowForm
	return f
}