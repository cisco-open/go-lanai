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

package actuator

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/access"
    "github.com/cisco-open/go-lanai/pkg/security/errorhandling"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2/tokenauth"
    matcherutils "github.com/cisco-open/go-lanai/pkg/utils/matcher"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/matcher"
    "net/http"
    "regexp"
)

/*******************************
	Interfaces
********************************/

// SecurityCustomizer is a single SecurityCustomizer can be registered via Registrar
// SecurityCustomizer is typically responsible to setup authentication scheme
// it should not configure access control, which is configured per-endpoint via properties or AccessControlCustomizer
type SecurityCustomizer interface {
	Customize(ws security.WebSecurity)
}

// SecurityCustomizerFunc convert a function to interface SecurityCustomizer
type SecurityCustomizerFunc func(ws security.WebSecurity)

func (f SecurityCustomizerFunc) Customize(ws security.WebSecurity) {
	f(ws)
}

// AccessControlCustomizer Similar to SecurityCustomizer, but is used to customize access control of each endpoint.
// Implementations of AccessControlCustomizer can be registered via Registrar, and NewAccessControlByPermissions is
// used if no other customizer is registered.
// Also See NewSimpleAccessControl, NewAccessControlByPermissions, NewAccessControlByScopes
type AccessControlCustomizer interface {
	Customize(ac *access.AccessControlFeature, epId string, paths []string)
}

// AccessControlCustomizeFunc convert a function to interface AccessControlCustomizer
type AccessControlCustomizeFunc func(ac *access.AccessControlFeature, epId string, paths []string)

func (f AccessControlCustomizeFunc) Customize(ac *access.AccessControlFeature, epId string, paths []string) {
	f(ac, epId, paths)
}

/*******************************
	Security Configurer
********************************/

// actuatorSecurityConfigurer implements security.Configurer
type actuatorSecurityConfigurer struct {
	properties   *ManagementProperties
	endpoints    WebEndpoints
	customizer   SecurityCustomizer
	acCustomizer AccessControlCustomizer
}

func newActuatorSecurityConfigurer(properties *ManagementProperties, endpoints WebEndpoints, customizer SecurityCustomizer, acCustomizer AccessControlCustomizer) *actuatorSecurityConfigurer {
	if customizer == nil {
		customizer = NewTokenAuthSecurity()
	}
	if acCustomizer == nil {
		acCustomizer = NewAccessControlByPermissions(properties.Security)
	}
	return &actuatorSecurityConfigurer{
		properties:   properties,
		endpoints:    endpoints,
		customizer:   customizer,
		acCustomizer: acCustomizer,
	}
}

func (c *actuatorSecurityConfigurer) Configure(ws security.WebSecurity) {
	if c.customizer != nil {
		c.customizer.Customize(ws)
	}

	path := fmt.Sprintf("%s/**", c.properties.Endpoints.Web.BasePath)

	ws.Route(matcher.RouteWithPattern(path).And(matcherutils.Not(matcher.RouteWithMethods(http.MethodOptions)))).
		With(errorhandling.New())

	// configure access control based on customizer and installed web endpoints
	ac := access.Configure(ws)
	for k, _ := range c.endpoints {
		c.acCustomizer.Customize(ac, k, c.endpoints.Paths(k))
	}

	// fallback configuration
	if c.properties.Security.EnabledByDefault {
		ac.Request(matcher.AnyRequest()).Authenticated()
	} else {
		ac.Request(matcher.AnyRequest()).PermitAll()
	}
}

/*******************************
	Common Implementation
********************************/

// NewTokenAuthSecurity returns a SecurityCustomizer config actuator security to use OAuth2 token auth.
// This is the default SecurityCustomizer if no other SecurityCustomizer is registered
func NewTokenAuthSecurity() SecurityCustomizer {
	return SecurityCustomizerFunc(func(ws security.WebSecurity) {
		ws.With(tokenauth.New())
	})
}


// NewSimpleAccessControl is a convenient AccessControlCustomizer constructor that create simple access
// control rule to ALL paths of each endpoint.
// A mapper function is required to convert each endpoint ID to its corresponding access.ControlFunc
func NewSimpleAccessControl(acCreator func(epId string) access.ControlFunc) AccessControlCustomizer {
	return AccessControlCustomizeFunc(func(ac *access.AccessControlFeature, epId string, paths []string) {
		if len(paths) == 0 {
			return
		}

		// configure request matchers
		m := pathToRequestPattern(paths[0])
		for _, p := range paths[1:] {
			m = m.Or(pathToRequestPattern(p))
		}

		// configure access control
		controlFunc := acCreator(epId)
		ac.Request(m).AllowIf(controlFunc)
	})
}

// NewAccessControlByPermissions returns a AccessControlCustomizer that uses SecurityProperties and given default
// permissions to setup access control of each endpoint.
// 1. If security of any particular endpoint is not enabled, access.PermitAll is used
// 2. If no permissions are configured in the properties and no defaults are given, access.Authenticated is used
//
// This is the default AccessControlCustomizer if no other AccessControlCustomizer is registered
func NewAccessControlByPermissions(properties SecurityProperties, defaultPerms ...string) AccessControlCustomizer {
	return NewSimpleAccessControl(func(epId string) access.ControlFunc {
		enabled, permissions := collectSecurityFacts(epId, &properties)
		if len(permissions) == 0 {
			permissions = defaultPerms
		}
		switch {
		case !enabled:
			return access.PermitAll
		case len(permissions) == 0:
			return access.Authenticated
		default:
			return access.HasPermissions(permissions...)
		}
	})
}

// NewAccessControlByScopes returns a AccessControlCustomizer that uses SecurityProperties and given default
// approved scopes to setup access control of each endpoint.
// "usePermissions" indicate if we should use permissions configured in SecurityProperties for scope checking
//
// 1. If security of any particular endpoint is not enabled, access.PermitAll is used
// 2. If usePermissions is true but no permissions are configured in SecurityProperties, defaultScopes is used to resolve scoes
// 3. If no scopes are configured (regardless if usePermissions is enabled), access.Authenticated is used
//
// Note: This customizer is particularly useful for client_credentials grant type
func NewAccessControlByScopes(properties SecurityProperties, usePermissions bool, defaultScopes ...string) AccessControlCustomizer {
	return NewSimpleAccessControl(func(epId string) access.ControlFunc {
		// first grab some facts
		scopes := defaultScopes
		enabled, permissions := collectSecurityFacts(epId, &properties)
		// if usePermissions is true, we use permissions from properties to for scope checking
		if usePermissions && len(permissions) != 0 {
			scopes = permissions
		}

		// then choose access control func
		switch {
		case !enabled:
			return access.PermitAll
		case len(scopes) == 0:
			return access.Authenticated
		default:
			return tokenauth.ScopesApproved(scopes...)
		}
	})
}

var pathVarRegex = regexp.MustCompile(`:[a-zA-Z0-9\-_]+`)

// pathToRequestPattern convert path variables to wildcard request pattern
// "/path/to/:any/endpoint" would converted to "/path/to/*/endpoint
func pathToRequestPattern(path string) web.RequestMatcher {
	patternStr := pathVarRegex.ReplaceAllString(path, "*")
	return matcher.RequestWithPattern(patternStr)
}

func collectSecurityFacts(epId string, properties *SecurityProperties, defaults ...string) (enabled bool, permissions []string) {
	permissions = defaults
	enabled = properties.EnabledByDefault
	if len(properties.Permissions) != 0 {
		permissions = properties.Permissions
	}
	if props, ok := properties.Endpoints[epId]; ok {
		permissions = props.Permissions
		if props.Enabled != nil {
			enabled = *props.Enabled
		}
	}
	return
}
