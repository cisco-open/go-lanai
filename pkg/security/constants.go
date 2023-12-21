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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
)

const (
	MinSecurityPrecedence = bootstrap.SecurityPrecedence
	MaxSecurityPrecedence = bootstrap.SecurityPrecedence + bootstrap.FrameworkModulePrecedenceBandwidth
)

const (
	ContextKeySecurity = web.ContextKeySecurity
)

const (
	// CompatibilityReference
	/**
	 * Note about compatibility reference:
	 *
	 * Whenever an incompatible security model changes (in terms of serialization) is made to the class,
	 * we should update the version tag.
	 *
	 * For now we use project version + incremental number as tag, but we could also use timestamp or date
	 */
	CompatibilityReference    = "4000"
	CompatibilityReferenceTag = "SMCR" // SMCR = Security Model Compatibility Ref
)

const (
	// SpecialPermissionAccessAllTenant
	// Deprecated: this permission is no longer sufficient to determine tenancy access
	// in the case of an oauth2 authentication where the client is also tenanted.
	// We are deprecating the use case where a user does not select a tenant.
	SpecialPermissionAccessAllTenant = "ACCESS_ALL_TENANTS"
	SpecialPermissionAPIAdmin        = "IS_API_ADMIN"
	SpecialPermissionSwitchTenant    = "SWITCH_TENANT"
	SpecialPermissionSwitchUser      = "VIEW_OPERATOR_LOGIN_AS_CUSTOMER"
	//SpecialPermissionAdmin = "IS_ADMIN"
	//SpecialPermissionOperator = "IS_OPERATOR"
	//SpecialPermission = ""
)

const (
	SpecialTenantIdWildcard = "*"
)

const (
	DetailsKeyAuthWarning = "AuthWarning"
	DetailsKeyAuthTime    = "AuthTime"
	DetailsKeyAuthMethod  = "AuthMethod"
	DetailsKeyMFAApplied  = "MFAApplied"
	DetailsKeySessionId   = "SessionId"
)

const (
	AuthMethodPassword       = "Password"
	AuthMethodExternalSaml   = "ExtSAML"
	AuthMethodExternalOpenID = "ExtOpenID"
)

const (
	WSSharedKeyCompositeAuthSuccessHandler  = "CompositeAuthSuccessHandler"
	WSSharedKeyCompositeAuthErrorHandler    = "CompositeAuthErrorHandler"
	WSSharedKeyCompositeAccessDeniedHandler = "CompositeAccessDeniedHandler"
	WSSharedKeySessionStore                 = "SessionStore"
	WSSharedKeyRequestPreProcessors         = "RequestPreProcessors"
)

// Middleware Orders
const (
	_ = HighestMiddlewareOrder + iota*20
	MWOrderSessionHandling
	MWOrderAuthPersistence
	MWOrderErrorHandling
	MWOrderCsrfHandling
	MWOrderOAuth2AuthValidation
	MWOrderSAMLMetadataRefresh
	MWOrderPreAuth
	MWOrderBasicAuth
	MWOrderFormLogout
	MWOrderFormAuth
	MWOrderOAuth2TokenAuth
	// ... more MW goes here
	MWOrderAccessControl     = LowestMiddlewareOrder - 200
	MWOrderOAuth2Endpoints   = MWOrderAccessControl + 100
	MWOrderSamlAuthEndpoints = MWOrderAccessControl + 100
)

// Feature Orders, if feature is not listed here, it's unordered. Unordered features are applied at last
const (
	_ = iota * 100
	FeatureOrderOAuth2ClientAuth
	FeatureOrderAuthenticator
	FeatureOrderBasicAuth
	FeatureOrderFormLogin
	FeatureOrderSamlLogin
	FeatureOrderSamlLogout
	FeatureOrderLogout
	FeatureOrderOAuth2TokenEndpoint
	FeatureOrderOAuth2AuthorizeEndpoint
	FeatureOrderSamlAuthorizeEndpoint
	FeatureOrderOAuth2TokenAuth
	FeatureOrderCsrf
	FeatureOrderAccess
	FeatureOrderSession
	FeatureOrderRequestCache
	// ... more Feature goes here
	FeatureOrderErrorHandling = order.Lowest - 200
)

// AuthenticationSuccessHandler Orders, if not listed here, it's unordered. Unordered handlers are applied at last
const (
	_                         = iota
	HandlerOrderChangeSession = iota * 100
	HandlerOrderConcurrentSession
)

// CSRF headers and parameter names - shared by CSRF feature and session feature's request cache
const (
	CsrfParamName  = "_csrf"
	CsrfHeaderName = "X-CSRF-TOKEN"
)
