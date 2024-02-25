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

package basicauth

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/errorhandling"
    "github.com/cisco-open/go-lanai/pkg/web/middleware"
)

var (
	FeatureId = security.FeatureId("BasicAuth", security.FeatureOrderBasicAuth)
)

// We currently don't have any stuff to configure
//goland:noinspection GoNameStartsWithPackageName
type BasicAuthFeature struct {
	entryPoint security.AuthenticationEntryPoint
}

// Standard security.Feature entrypoint
func (f *BasicAuthFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func Configure(ws security.WebSecurity) *BasicAuthFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*BasicAuthFeature)
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *BasicAuthFeature {
	return &BasicAuthFeature{
		entryPoint: NewBasicAuthEntryPoint(),
	}
}

func (f *BasicAuthFeature) EntryPoint(entrypoint security.AuthenticationEntryPoint) *BasicAuthFeature {
	f.entryPoint = entrypoint
	return f
}

//goland:noinspection GoNameStartsWithPackageName
type BasicAuthConfigurer struct {

}

func newBasicAuthConfigurer() *BasicAuthConfigurer {
	return &BasicAuthConfigurer{
	}
}

func (bac *BasicAuthConfigurer) Apply(f security.Feature, ws security.WebSecurity) error {

	// additional error handling
	errorHandler := ws.Shared(security.WSSharedKeyCompositeAuthErrorHandler).(*security.CompositeAuthenticationErrorHandler)
	errorHandler.Add(NewBasicAuthErrorHandler())

	// default is NewBasicAuthEntryPoint(). But security.Configurer have chance to overwrite it or unset it
	if entrypoint := f.(*BasicAuthFeature).entryPoint; entrypoint != nil {
		errorhandling.Configure(ws).
			AuthenticationEntryPoint(entrypoint)
	}

	// configure middlewares
	basicAuth := NewBasicAuthMiddleware(
		ws.Authenticator(),
		ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(security.AuthenticationSuccessHandler),
		)

	auth := middleware.NewBuilder("basic auth").
		Order(security.MWOrderBasicAuth).
		Use(basicAuth.HandlerFunc())

	ws.Add(auth)
	return nil
}