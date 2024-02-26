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
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/gin-gonic/gin"
    "net/http"
)

/***************************************
	Additional Context for Internal
****************************************/

// FeatureModifier add or remove features. \
// Should not used directly by service
// use corresponding feature's Configure(WebSecurity) instead
type FeatureModifier interface {
	// Enable kick off configuration of give Feature.
	// If the given Feature is not enabled yet, it's added to the receiver and returned
	// If the given Feature is already enabled, the already enabled Feature is returned
	Enable(Feature) Feature
	// Disable remove given feature using its FeatureIdentifier
	Disable(Feature)
}

type WebSecurityReader interface {
	GetRoute() web.RouteMatcher
	GetCondition() web.RequestMatcher
	GetHandlers() []interface{}
}

type WebSecurityMappingBuilder interface {
	Build() []web.Mapping
}

// FeatureConfigurer not intended to be used directly in service
type FeatureConfigurer interface {
	Apply(Feature, WebSecurity) error
}

type FeatureRegistrar interface {
	// RegisterFeature is typically used by feature packages, such as session, oauth, etc
	// not intended to be used directly in service
	RegisterFeature(featureId FeatureIdentifier, featureConfigurer FeatureConfigurer)

	// FindFeature is typically used by feature packages
	FindFeature(featureId FeatureIdentifier) FeatureConfigurer
}

func NoopHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = c.AbortWithError(http.StatusNotFound, fmt.Errorf("page not found"))
	}
}
