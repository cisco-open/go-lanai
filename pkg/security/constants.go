package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
)

const (
	MinSecurityPrecedence = bootstrap.FrameworkModulePrecedence + 2000
	MaxSecurityPrecedence = bootstrap.FrameworkModulePrecedence + 2999


	ContextKeySecurity = web.ContextKeySecurity
)

// Middleware Orders
const (
	_ = iota
	MWOrderSessionHandling = HighestMiddlewareOrder + iota * 20
	MWOrderErrorHandling
	MWOrderAuthPersistence
	MWOrderBasicAuth
	// ... TODO more MW goes here
	MWOrderAccessControl = LowestMiddlewareOrder - 200
)

// Feature Orders, if feature is not listed here, it's unordered. Unordered features are applied at last
const (
	_ = iota
	FeatureOrderAuthenticator = order.Highest + iota * 100
	FeatureOrderBasicAuth
	// ... TODO more Feature goes here
	FeatureOrderErrorHandling = order.Lowest - 200
)
