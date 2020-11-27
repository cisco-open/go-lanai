package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
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

// MiddlewareCondition accept *http.Request and can be translated to web.MWConditionFunc
type MiddlewareCondition matcher.ChainableMatcher

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

	// Route configure the path and method pattern which this WebSecurity applies to
	Route(web.RouteMatcher) WebSecurity

	// Condition sets additional conditions of incoming request which this WebSecurity applies to
	Condition(mwcm web.MWConditionMatcher) WebSecurity

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
type simpleFeatureId string
// FeatureIdentifier interface
func (id simpleFeatureId) String() string {
	return string(id)
}

// FeatureIdentifier interface
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

// order.Ordered interface
func (id featureId) Order() int {
	return id.order
}

// FeatureIdentifier interface
func (id featureId) String() string {
	return id.id
}

// FeatureIdentifier interface
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

// order.PriorityOrdered interface
func (id priorityFeatureId) PriorityOrder() int {
	return id.order
}

// FeatureIdentifier interface
func (id priorityFeatureId) String() string {
	return id.id
}

// FeatureIdentifier interface
func (id priorityFeatureId) GoString() string {
	return id.id
}

// PriorityFeatureId create an priority ordered FeatureIdentifier
func PriorityFeatureId(id string, order int) FeatureIdentifier {
	return priorityFeatureId{id: id, order: order}
}

