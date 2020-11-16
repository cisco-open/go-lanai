package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
)

/***************************************
	Additional Context for Internal
****************************************/
// FeatureModifier add or remove features. \
// Should not used directly by service
// use corresponding feature's Configure(WebSecurity) instead
type FeatureModifier interface {
	Enable(Feature) error
	Disable(Feature)
}

type WebSecurityMiddlewareBuilder interface {
	Build() []web.MiddlewareMapping
}

// FeatureConfigurer not intended to be used directly in service
type FeatureConfigurer interface {
	Apply(Feature, WebSecurity) error
}

type FeatureRegistrar interface {
	// RegisterFeature is typically used by feature packages, such as session, oauth, etc
	// not intended to be used directly in service
	//
	RegisterFeature(featureId FeatureIdentifier, featureConfigurer FeatureConfigurer)
}
