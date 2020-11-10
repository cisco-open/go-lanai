package init

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"reflect"
)

const (
	MWOrderBasicAuth = security.HighestMiddlewareOrder + 200
)

var BasicAuthConfigurerType = reflect.TypeOf((*BasicAuthConfigurer)(nil))

// We currently don't have any stuff to configure
type BasicAuthFeature struct {
	// TODO we may want to override authenticator and other stuff
}

func (f *BasicAuthFeature) ConfigurerType() reflect.Type {
	return BasicAuthConfigurerType
}

func Configure(ws security.WebSecurity) *BasicAuthFeature {
	feature := &BasicAuthFeature{}
	if fc, ok := ws.(security.FeatureModifier); ok {
		_ = fc.Enable(feature) // we ignore error here
		return feature
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

type BasicAuthConfigurer struct {
	authenticator security.Authenticator
}

func newBasicAuthConfigurer(auth security.Authenticator) *BasicAuthConfigurer {
	return &BasicAuthConfigurer{
		authenticator: auth,
	}
}

func (bac *BasicAuthConfigurer) Build(_ security.Feature) ([]security.MiddlewareTemplate, error) {
	// TODO
	basicAuth := basicauth.NewBasicAuthMiddleware(bac.authenticator)
	auth := middleware.NewBuilder("basic auth").
		Order(MWOrderBasicAuth).
		Use(basicAuth.HandlerFunc())

	return []security.MiddlewareTemplate{auth}, nil
}