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

package errorhandling

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/log"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/web/middleware"
)

var (
	FeatureId       = security.FeatureId("ErrorHandling", security.FeatureOrderErrorHandling)
)

// We currently don't have any stuff to configure
//goland:noinspection GoNameStartsWithPackageName
type ErrorHandlingFeature struct {
	authEntryPoint      security.AuthenticationEntryPoint
	accessDeniedHandler security.AccessDeniedHandler
	authErrorHandler    security.AuthenticationErrorHandler
	errorHandler        *security.CompositeErrorHandler
}

// Standard security.Feature entrypoint
func (f *ErrorHandlingFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func (f *ErrorHandlingFeature) AuthenticationEntryPoint(v security.AuthenticationEntryPoint) *ErrorHandlingFeature {
	f.authEntryPoint = v
	return f
}

func (f *ErrorHandlingFeature) AccessDeniedHandler(v security.AccessDeniedHandler) *ErrorHandlingFeature {
	f.accessDeniedHandler = v
	return f
}

func (f *ErrorHandlingFeature) AuthenticationErrorHandler(v security.AuthenticationErrorHandler) *ErrorHandlingFeature {
	f.authErrorHandler = v
	return f
}

// AdditionalErrorHandler add security.ErrorHandler to existing list.
// This value is typically used by other features, because there are no other type of concrete errors except for
// AuthenticationError and AccessControlError,
// which are handled by AccessDeniedHandler, AuthenticationErrorHandler and AuthenticationEntryPoint
func (f *ErrorHandlingFeature) AdditionalErrorHandler(v security.ErrorHandler) *ErrorHandlingFeature {
	f.errorHandler.Add(v)
	return f
}

func Configure(ws security.WebSecurity) *ErrorHandlingFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*ErrorHandlingFeature)
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *ErrorHandlingFeature {
	return &ErrorHandlingFeature{
		errorHandler: security.NewErrorHandler(),
	}
}

//goland:noinspection GoNameStartsWithPackageName
type ErrorHandlingConfigurer struct {

}

func newErrorHandlingConfigurer() *ErrorHandlingConfigurer {
	return &ErrorHandlingConfigurer{
	}
}

func (ehc *ErrorHandlingConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	// Verify
	if err := ehc.validate(feature.(*ErrorHandlingFeature), ws); err != nil {
		return err
	}
	f := feature.(*ErrorHandlingFeature)

	authErrorHandler := ws.Shared(security.WSSharedKeyCompositeAuthErrorHandler).(*security.CompositeAuthenticationErrorHandler)
	authErrorHandler.Add(f.authErrorHandler)

	accessDeniedHandler := ws.Shared(security.WSSharedKeyCompositeAccessDeniedHandler).(*security.CompositeAccessDeniedHandler)
	accessDeniedHandler.Add(f.accessDeniedHandler)

	mw := NewErrorHandlingMiddleware()
	mw.entryPoint = f.authEntryPoint
	mw.accessDeniedHandler = accessDeniedHandler
	mw.authErrorHandler = authErrorHandler
	mw.errorHandler = f.errorHandler

	errHandler := middleware.NewBuilder("error handling").
		Order(security.MWOrderErrorHandling).
		Use(mw.HandlerFunc())

	ws.Add(errHandler)
	return nil
}


func (ehc *ErrorHandlingConfigurer) validate(f *ErrorHandlingFeature, ws security.WebSecurity) error {
	if f.authEntryPoint == nil {
		logger.WithContext(ws.Context()).Infof("authentication entry point is not set, fallback to access denied handler - [%v], ", log.Capped(ws, 80))
	}

	if f.authErrorHandler == nil {
		logger.WithContext(ws.Context()).Infof("using default authentication error handler - [%v]", log.Capped(ws, 80))
		f.authErrorHandler = &security.DefaultAuthenticationErrorHandler{}
	}

	if f.accessDeniedHandler == nil {
		logger.WithContext(ws.Context()).Infof("using default access denied handler - [%v]", log.Capped(ws, 80))
		f.accessDeniedHandler = &security.DefaultAccessDeniedHandler{}
	}

	// always add a default to the end. note: DefaultErrorHandler is unordered
	f.errorHandler.Add(&security.DefaultErrorHandler{})
	return nil
}

