package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
)

const (
	WSSharedKeyAuthenticatorManager = "kAuthenticator"
)
/************************************
	Interfaces for setting security
*************************************/
// Configurer can be registered to Registrar.
// Each Configurer will get a newly created WebSecurity and is responsible to configure for customized security
type Configurer interface {
	Configure(WebSecurity)
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
	// Register is the entry point for all security configuration.
	// Microservice or other library packages typically call this in Provide or Invoke functions
	// Note: use this function inside fx.Lifecycle takes no effect
	Initialize(registrar *web.Registrar) error
}

/****************************************
	Type definitions for
    specifying web security specs
*****************************************/
// MiddlewareTemplate is partially configured middleware.MappingBuilder.
// it holds the middleware's gin.HandlerFunc and its order
type MiddlewareTemplate *middleware.MappingBuilder

// FeatureIdentifier is unique for each feature.
// Security initializer use this value to locate corresponding FeatureConfigurer
// or sort configuration order
type FeatureIdentifier interface{}

// Feature holds security settings of specific feature.
// Any Feature should have a corresponding FeatureConfigurer
type Feature interface {
	Identifier() FeatureIdentifier
}

// WebSecurity holds information on security configuration
type WebSecurity interface {

	ApplyTo(r web.RouteMatcher) WebSecurity

	// Add is DSL style setter to add MiddlewareTemplate
	Add(...MiddlewareTemplate) WebSecurity

	// Remove is DSL style setter to add remove MiddlewareTemplate
	Remove(...MiddlewareTemplate) WebSecurity

	// With is DSL style setter to enable features
	With(f Feature) WebSecurity

	// Shared get shared value
	Shared(key string) interface{}

	// AddShared add shared value. returns error when the key already exists
	AddShared(key string, value interface{}) error

	// returns Authenticator
	Authenticator() Authenticator

	// Features get currently configured Feature list
	Features() []Feature
}

/****************************************
	Convenient Types
*****************************************/


