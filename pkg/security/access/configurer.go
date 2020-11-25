package access

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
)

const (
	MWOrderAccessControl = security.LowestMiddlewareOrder - 200
	FeatureId = "AC"
)

//goland:noinspection GoNameStartsWithPackageName
type AccessControlFeature struct {
	// TODO we may want to override authenticator and other stuff
}

// Standard security.Feature entrypoint
func (f *AccessControlFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func Configure(ws security.WebSecurity) *AccessControlFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		_ = fc.Enable(feature) // we ignore error here
		return feature
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *AccessControlFeature {
	return &AccessControlFeature{}
}

//goland:noinspection GoNameStartsWithPackageName
type AccessControlConfigurer struct {

}

func newAccessControlConfigurer() *AccessControlConfigurer {
	return &AccessControlConfigurer{
	}
}

func (bac *AccessControlConfigurer) Apply(_ security.Feature, ws security.WebSecurity) error {
	// TODO
	mw := NewAccessControlMiddleware()
	ac := middleware.NewBuilder("access control").
		Order(MWOrderAccessControl).
		Use(mw.ACHandlerFunc())

	ws.Add(ac)
	return nil
}