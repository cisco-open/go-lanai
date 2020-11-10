package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"reflect"
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

// Feature holds security settings of specific feature.
type Feature interface {
	ConfigurerType() reflect.Type
}

// WebSecurity holds information on security configuration
type WebSecurity interface {
	ApplyTo(r web.RouteMatcher) WebSecurity
	Add(MiddlewareTemplate) WebSecurity
	Remove(MiddlewareTemplate) WebSecurity
	Features() []Feature
}




