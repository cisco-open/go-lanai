package serviceinit

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/resserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"go.uber.org/fx"
)

// newResServerConfigurer required for token auth
func newResServerConfigurer() resserver.ResourceServerConfigurer {
	return func(config *resserver.Configuration) {
		//do nothing
	}
}

type secDI struct {
	fx.In
	SecReg         security.Registrar
	ActrReg        *actuator.Registrar           `optional:"true"`
	ActrProperties actuator.ManagementProperties `optional:"true"`
	HealthReg      health.Registrar              `optional:"true"`
}

// healthDisclosureControl is a custom health details disclosure control.
// This example allows all users to see health details.
// TODO implement this properly for desired security model
func healthDisclosureControl() health.DisclosureControlFunc {
	return func(ctx context.Context) bool {
		return true
	}
}

// TODO implement this properly for desired security model
func configureSecurity(di secDI) {
	// Configure custom security of actuator endpoint here, if applicable.
	// This example doesn't setup any custom security for actuator. Everything is configured via application.yml
	if di.ActrReg != nil {
		//acCustomizer := actuator.NewAccessControlByScopes(di.ActrProperties.Security, true, service.SpecialScopeAdmin)
		//di.ActrReg.MustRegister(acCustomizer)
	}

	// Configure how health details is disclosed.
	// This example doesn't setup any custom logic. Everything is configured via application.yml
	if di.HealthReg != nil {
		//di.HealthReg.MustRegister(healthDisclosureControl())
	}

	// Setup API security
	di.SecReg.Register(&securityConfigurer{})
}

// security configuration for APIs.
// This example enable token authentication for all APIs, and allow access for any authenticated user
type securityConfigurer struct{}

func (c *securityConfigurer) Configure(ws security.WebSecurity) {
	// DSL style example
	// for REST API
	ws.Route(matcher.RouteWithPattern("/api/**")).
		With(tokenauth.New()).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(errorhandling.New())
}
