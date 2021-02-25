package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
)

const (
	MinSecurityPrecedence = bootstrap.SecurityPrecedence
	MaxSecurityPrecedence = bootstrap.SecurityPrecedence + bootstrap.FrameworkModulePrecedenceBandwith
)

const (
	ContextKeySecurity = web.ContextKeySecurity
)

const (
	/**
	 * Note about compatibility reference:
	 *
	 * Whenever an incompatible security model changes (in terms of serialization) is made to the class,
	 * we should update the version tag.
	 *
	 * For now we use project version + incremental number as tag, but we could also use timestamp or date
	 */
	CompatibilityReference = "4000"
	CompatibilityReferenceTag = "SMCR" // SMCR = Security Model Compatibility Ref
)

const (
	SpecialPermissionAccessAllTenant = "ACCESS_ALL_TENANTS"
	SpecialPermissionAPIAdmin = "IS_API_ADMIN"
	SpecialPermissionSwitchTenant = "SWITCH_TENANT"
	SpecialPermissionSwitchUser = "VIEW_OPERATOR_LOGIN_AS_CUSTOMER"
	//SpecialPermissionAdmin = "IS_ADMIN"
	//SpecialPermissionOperator = "IS_OPERATOR"
	//SpecialPermission = ""
)

const (
	DetailsKeyAuthWarning = "AuthWarning"
	DetailsKeyAuthTime    = "AuthTime"
)

const (
	WSSharedKeyCompositeAuthSuccessHandler = "CompositeAuthSuccessHandler"
	WSSharedKeyCompositeAuthErrorHandler = "CompositeAuthErrorHandler"
	WSSharedKeyCompositeAccessDeniedHandler = "CompositeAccessDeniedHandler"
	WSSharedKeySessionStore = "SessionStore"
	WSSharedKeyRequestPreProcessors = "RequestPreProcessors"
)

// Middleware Orders
const (
	_ = HighestMiddlewareOrder + iota * 20
	MWOrderSessionHandling
	MWOrderAuthPersistence
	MWOrderErrorHandling
	MWOrderCsrfHandling
	MWOrderOAuth2AuthValidation
	MWOrderPreAuth
	MWOrderBasicAuth
	MWOrderFormLogout
	MWOrderFormAuth
	MWOrderOAuth2TokenAuth
	MWOrderSAMLMetadataRefresh
	// ... TODO more MW goes here
	MWOrderAccessControl = LowestMiddlewareOrder - 200
	MWOrderOAuth2Endpoints = MWOrderAccessControl + 100
)

// Feature Orders, if feature is not listed here, it's unordered. Unordered features are applied at last
const (
	_ = iota * 100
	FeatureOrderOAuth2ClientAuth
	FeatureOrderAuthenticator
	FeatureOrderBasicAuth
	FeatureOrderFormLogin
	FeatureOrderSamlLogin
	FeatureOrderLogout
	FeatureOrderOAuth2TokenEndpoint
	FeatureOrderOAuth2AuthorizeEndpoint
	FeatureOrderOAuth2TokenAuth
	FeatureOrderCsrf
	FeatureOrderAccess
	FeatureOrderSession
	FeatureOrderRequestCache
	// ... TODO more Feature goes here
	FeatureOrderErrorHandling = order.Lowest - 200
)

// AuthenticationSuccessHandler Orders, if not listed here, it's unordered. Unordered handlers are applied at last
const (
	_ = iota
	HandlerOrderChangeSession = iota * 100
	HandlerOrderConcurrentSession

)

// CSRF headers and parameter names - shared by CSRF feature and session feature's request cache
const (
	CsrfParamName  = "_csrf"
	CsrfHeaderName = "X-CSRF-TOKEN"
)
