package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"reflect"
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
	Build(Feature) ([]MiddlewareTemplate, error)
}


type FeatureRegistrar interface {
	// RegisterFeatureConfigurer is typically used by feature packages, such as session, oauth, etc
	// not intended to be used directly in service
	RegisterFeatureConfigurer(reflect.Type, FeatureConfigurer)
}
