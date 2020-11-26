package security

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"

const (
	MinSecurityPrecedence = bootstrap.FrameworkModulePrecedence + 2000
	MaxSecurityPrecedence = bootstrap.FrameworkModulePrecedence + 2999

	ContextKeySecurity = "kSecurity"

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
