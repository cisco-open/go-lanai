package basicauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
)

var (
	FeatureId = security.FeatureId("BasicAuth", security.FeatureOrderBasicAuth)
)

// We currently don't have any stuff to configure
//goland:noinspection GoNameStartsWithPackageName
type BasicAuthFeature struct {
	entryPoint security.AuthenticationEntryPoint
}

// Standard security.Feature entrypoint
func (f *BasicAuthFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func Configure(ws security.WebSecurity) *BasicAuthFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*BasicAuthFeature)
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *BasicAuthFeature {
	return &BasicAuthFeature{
		entryPoint: NewBasicAuthEntryPoint(),
	}
}

func (f *BasicAuthFeature) EntryPoint(entrypoint security.AuthenticationEntryPoint) *BasicAuthFeature {
	f.entryPoint = entrypoint
	return f
}

//goland:noinspection GoNameStartsWithPackageName
type BasicAuthConfigurer struct {

}

func newBasicAuthConfigurer() *BasicAuthConfigurer {
	return &BasicAuthConfigurer{
	}
}

func (bac *BasicAuthConfigurer) Apply(f security.Feature, ws security.WebSecurity) error {

	// additional error handling
	errorHandler := ws.Shared(security.WSSharedKeyCompositeAuthErrorHandler).(*security.CompositeAuthenticationErrorHandler)
	errorHandler.Add(NewBasicAuthErrorHandler())

	// default is NewBasicAuthEntryPoint(). But security.Configurer have chance to overwrite it or unset it
	if entrypoint := f.(*BasicAuthFeature).entryPoint; entrypoint != nil {
		errorhandling.Configure(ws).
			AuthenticationEntryPoint(entrypoint)
	}

	// configure middlewares
	basicAuth := NewBasicAuthMiddleware(
		ws.Authenticator(),
		ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(security.AuthenticationSuccessHandler),
		)

	auth := middleware.NewBuilder("basic auth").
		Order(security.MWOrderBasicAuth).
		Use(basicAuth.HandlerFunc())

	ws.Add(auth)
	return nil
}