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

package samlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"net/url"
)

var (
	FeatureId    = security.FeatureId("SamlAuthorizeEndpoint", security.FeatureOrderSamlAuthorizeEndpoint)
	SloFeatureId = security.FeatureId("SamlSLOEndpoint", security.FeatureOrderSamlLogout)
)

type Feature struct {
	id            security.FeatureIdentifier
	ssoCondition  web.RequestMatcher
	ssoLocation   *url.URL
	signingMethod string
	metadataPath  string
	issuer        security.Issuer
	logoutUrl     string
}

// New Standard security.Feature entrypoint for authorization, DSL style. Used with security.WebSecurity
func New() *Feature {
	return &Feature{
		id: FeatureId,
	}
}

// NewLogout Standard security.Feature entrypoint for single-logout, DSL style. Used with security.WebSecurity
func NewLogout() *Feature {
	return &Feature{
		id: SloFeatureId,
	}
}

func (f *Feature) Identifier() security.FeatureIdentifier {
	return f.id
}

func (f *Feature) SsoCondition(condition web.RequestMatcher) *Feature {
	f.ssoCondition = condition
	return f
}

func (f *Feature) SsoLocation(location *url.URL) *Feature {
	f.ssoLocation = location
	return f
}

func (f *Feature) MetadataPath(path string) *Feature {
	f.metadataPath = path
	return f
}

func (f *Feature) Issuer(issuer security.Issuer) *Feature {
	f.issuer = issuer
	return f
}

func (f *Feature) SigningMethod(signatureMethod string) *Feature {
	f.signingMethod = signatureMethod
	return f
}

// EnableSLO when logoutUrl is set, SLO Request handling is added to logout.Feature.
// SLO feature cannot work properly if this value mismatches the logout URL
func (f *Feature) EnableSLO(logoutUrl string) *Feature {
	f.logoutUrl = logoutUrl
	return f
}

func Configure(ws security.WebSecurity) *Feature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*Feature)
	}
	panic(fmt.Errorf("unable to configure saml authserver: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

func ConfigureLogout(ws security.WebSecurity) *Feature {
	feature := NewLogout()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*Feature)
	}
	panic(fmt.Errorf("unable to configure saml authserver: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}
