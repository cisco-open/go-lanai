package init

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
)

const (
	MWOrderBasicAuth = security.HighestMiddlewareOrder + 200
	BasicFeatureId = "BasicAuth"
)

// We currently don't have any stuff to configure
type BasicAuthFeature struct {
	// TODO we may want to override authenticator and other stuff
}

func (f *BasicAuthFeature) Identifier() security.FeatureIdentifier {
	return BasicFeatureId
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

}

func newBasicAuthConfigurer() *BasicAuthConfigurer {
	return &BasicAuthConfigurer{
	}
}

func (bac *BasicAuthConfigurer) Apply(_ security.Feature, ws security.WebSecurity) error {
	// TODO
	basicAuth := basicauth.NewBasicAuthMiddleware(ws.Authenticator())
	auth := middleware.NewBuilder("basic auth").
		Order(MWOrderBasicAuth).
		Use(basicAuth.HandlerFunc())

	ws.Add(auth)
	return nil
}