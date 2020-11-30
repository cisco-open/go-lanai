package access

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"fmt"
)

var (
	FeatureId = security.SimpleFeatureId("AC")
)

//goland:noinspection GoNameStartsWithPackageName
type AccessControlConfigurer struct {

}

func newAccessControlConfigurer() *AccessControlConfigurer {
	return &AccessControlConfigurer{
	}
}

func (acc *AccessControlConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	// Validate
	if err := acc.validate(feature.(*AccessControlFeature), ws); err != nil {
		return err
	}
	f := feature.(*AccessControlFeature)

	// construct decision maker functions
	decisionMakers := make([]DecisionMakerFunc, len(f.acl))
	for i,ac := range f.acl {
		decisionMakers[i] = MakeDecisionMakerFunc(ac.matcher, ac.control)
	}

	// register middlewares
	mw := NewAccessControlMiddleware(decisionMakers...)
	ac := middleware.NewBuilder("access control").
		Order(security.MWOrderAccessControl).
		Use(mw.ACHandlerFunc())

	ws.Add(ac)
	return nil
}

func (acc *AccessControlConfigurer) validate(f *AccessControlFeature, ws security.WebSecurity) error {
	if len(f.acl) == 0 {
		fmt.Printf("access control for routes match [%v] is not set. Default to DenyAll", ws)
		f.Request(matcher.AnyRequest()).DenyAll()
	}
	return nil
}
