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

package token

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"fmt"
)

// We currently don't have any stuff to configure
//goland:noinspection GoNameStartsWithPackageName
type TokenFeature struct {
	path string
	granters []auth.TokenGranter
}

// Standard security.Feature entrypoint
func (f *TokenFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func Configure(ws security.WebSecurity) *TokenFeature {
	feature := NewEndpoint()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*TokenFeature)
	}
	panic(fmt.Errorf("unable to configure oauth2 authserver: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func NewEndpoint() *TokenFeature {
	return &TokenFeature{
	}
}

/** Setters **/
func (f *TokenFeature) Path(path string) *TokenFeature {
	f.path = path
	return f
}

func (f *TokenFeature) AddGranter(granter auth.TokenGranter) *TokenFeature {
	if composite, ok := granter.(*auth.CompositeTokenGranter); ok {
		f.granters = append(f.granters, composite.Delegates()...)
	} else {
		f.granters = append(f.granters, granter)
	}

	return f
}