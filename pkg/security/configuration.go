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

package security

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/mapping"
    "github.com/cisco-open/go-lanai/pkg/web/middleware"
    "go.uber.org/fx"
)

/************************************
	Interfaces for setting security
*************************************/

// Configurer can be registered to Registrar.
// Each Configurer will get a newly created WebSecurity and is responsible to configure for customized security
type Configurer interface {
	Configure(WebSecurity)
}

type ConfigurerFunc func(ws WebSecurity)

func (f ConfigurerFunc) Configure(ws WebSecurity) {
	f(ws)
}

/************************************
	Interfaces for other modules
*************************************/

// Registrar is the entry point to configure security
type Registrar interface {
	// Register is the entry point for all security configuration.
	// Microservice or other library packages typically call this in Provide or Invoke functions
	// Note: use this function inside fx.Lifecycle takes no effect
	Register(...Configurer)
}

// Initializer is the entry point to bootstrap security
type Initializer interface {
	// Initialize is the entry point for all security configuration.
	// Microservice or other library packages typically call this in Provide or Invoke functions
	// Note: use this function inside fx.Lifecycle takes no effect
	Initialize(ctx context.Context, lc fx.Lifecycle, registrar *web.Registrar) error
}

/****************************************
	Type definitions for
    specifying web security specs
*****************************************/

// MiddlewareTemplate is partially configured middleware.MappingBuilder.
// it holds the middleware's gin.HandlerFunc and order
// if its route matcher and condition is not set, WebSecurity would make it matches WebSecurity's own values
type MiddlewareTemplate *middleware.MappingBuilder

// SimpleMappingTemplate is partially configured mapping.MappingBuilder
// it holds the simple mapping's path, gin.HandlerFunc and order
// if its condition is not set, WebSecurity would make it matches WebSecurity's own values
type SimpleMappingTemplate *mapping.MappingBuilder

// FeatureIdentifier is unique for each feature.
// Security initializer use this value to locate corresponding FeatureConfigurer
// or sort configuration order
type FeatureIdentifier interface {
	fmt.Stringer
	fmt.GoStringer
}

// Feature holds security settings of specific feature.
// Any Feature should have a corresponding FeatureConfigurer
type Feature interface {
	Identifier() FeatureIdentifier
}

// WebSecurity holds information on security configuration
type WebSecurity interface {

	// Context returns the context associated with the WebSecurity.
	// It's typlically holds bootstrap.ApplicationContext or its derived context
	// this should not returns nil
	Context() context.Context

	// Route configure the path and method pattern which this WebSecurity applies to
	// Calling this method multiple times concatenate all given matchers with OR operator
	Route(web.RouteMatcher) WebSecurity

	// Condition sets additional conditions of incoming request which this WebSecurity applies to
	// Calling this method multiple times concatenate all given matchers with OR operator
	Condition(mwcm web.RequestMatcher) WebSecurity

	// AndCondition sets additional conditions of incoming request which this WebSecurity applies to
	// Calling this method multiple times concatenate all given matchers with AND operator
	AndCondition(mwcm web.RequestMatcher) WebSecurity

	// Add is DSL style setter to add:
	// - MiddlewareTemplate
	// - web.MiddlewareMapping
	// - web.MvcMapping
	// - web.StaticMapping
	// - web.SimpleMapping
	// when MiddlewareTemplate is given, WebSecurity's Route and Condition are applied to it
	// this method panic if other type is given
	Add(...interface{}) WebSecurity

	// Remove is DSL style setter to remove:
	// - MiddlewareTemplate
	// - web.MiddlewareMapping
	// - web.MvcMapping
	// - web.StaticMapping
	// - web.SimpleMapping
	Remove(...interface{}) WebSecurity

	// With is DSL style setter to enable features
	With(f Feature) WebSecurity

	// Shared get shared value
	Shared(key string) interface{}

	// AddShared add shared value. returns error when the key already exists
	AddShared(key string, value interface{}) error

	// Authenticator returns Authenticator
	Authenticator() Authenticator

	// Features get currently configured Feature list
	Features() []Feature
}

/****************************************
	Convenient Types
*****************************************/

type simpleFeatureId string

// String implements FeatureIdentifier interface
func (id simpleFeatureId) String() string {
	return string(id)
}

// GoString implements FeatureIdentifier interface
func (id simpleFeatureId) GoString() string {
	return string(id)
}

// SimpleFeatureId create unordered FeatureIdentifier
func SimpleFeatureId(id string) FeatureIdentifier {
	return simpleFeatureId(id)
}

// featureId is ordered
type featureId struct {
	id string
	order int
}

// Order implements order.Ordered interface
func (id featureId) Order() int {
	return id.order
}

// String implements FeatureIdentifier interface
func (id featureId) String() string {
	return id.id
}

// GoString implements FeatureIdentifier interface
func (id featureId) GoString() string {
	return id.id
}

// FeatureId create an ordered FeatureIdentifier
func FeatureId(id string, order int) FeatureIdentifier {
	return featureId{id: id, order: order}
}

// priorityFeatureId is priority Ordered
type priorityFeatureId struct {
	id string
	order int
}

// PriorityOrder implements order.PriorityOrdered interface
func (id priorityFeatureId) PriorityOrder() int {
	return id.order
}

// String implements FeatureIdentifier interface
func (id priorityFeatureId) String() string {
	return id.id
}

// GoString implements FeatureIdentifier interface
func (id priorityFeatureId) GoString() string {
	return id.id
}

// PriorityFeatureId create a priority ordered FeatureIdentifier
func PriorityFeatureId(id string, order int) FeatureIdentifier {
	return priorityFeatureId{id: id, order: order}
}

