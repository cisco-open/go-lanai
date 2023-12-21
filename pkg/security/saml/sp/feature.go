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

package sp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
)

var (
	FeatureId       = security.FeatureId("saml_login", security.FeatureOrderSamlLogin)
	LogoutFeatureId = security.FeatureId("saml_logout", security.FeatureOrderSamlLogout)
)

type Feature struct {
	id             security.FeatureIdentifier
	metadataPath   string
	acsPath        string
	sloPath        string
	errorPath      string //The path to send the user to when authentication error is encountered
	successHandler security.AuthenticationSuccessHandler
	issuer         security.Issuer
}

func new(id security.FeatureIdentifier) *Feature {
	return &Feature{
		id:           id,
		metadataPath: "/saml/metadata",
		acsPath:      "/saml/SSO",
		sloPath:      "/saml/slo",
		errorPath:    "/error",
	}
}

func New() *Feature {
	return new(FeatureId)
}

func NewLogout() *Feature {
	return new(LogoutFeatureId)
}

func (f *Feature) Identifier() security.FeatureIdentifier {
	return f.id
}

func (f *Feature) Issuer(issuer security.Issuer) *Feature {
	f.issuer = issuer
	return f
}

func (f *Feature) ErrorPath(path string) *Feature {
	f.errorPath = path
	return f
}
