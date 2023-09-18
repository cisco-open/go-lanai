package serviceinit

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	opaaccess "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/access"
	opaactuator "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/resserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"go.uber.org/fx"
)

// newResServerConfigurer required for token auth
func newResServerConfigurer() resserver.ResourceServerConfigurer {
	return func(config *resserver.Configuration) {
		//do nothing
	}
}

type securityConfigurer struct{}

func (c *securityConfigurer) Configure(ws security.WebSecurity) {

	// DSL style example
	// for REST API
	ws.Route(matcher.RouteWithPattern("/api/**")).
		//Condition(matcher.RequestWithHost("localhost:8080")).
		With(tokenauth.New()).
		With(access.New().
			Request(matcher.AnyRequest()).
			WithOrder(order.Highest).
			// TODO Verify if policy path is correct
			CustomDecisionMaker(opaaccess.DecisionMakerWithOPA(opa.RequestQueryWithPolicy("testservice/allow_api"))),
		).
		With(errorhandling.New())
}

type secDI struct {
	fx.In
	SecReg         security.Registrar
	ActrReg        *actuator.Registrar           `optional:"true"`
	ActrProperties actuator.ManagementProperties `optional:"true"`
	HealthReg      health.Registrar              `optional:"true"`
}

func configureSecurity(di secDI) {
	if di.ActrReg != nil {
		acCustomizer := opaactuator.NewAccessControlWithOPA(di.ActrProperties.Security, opa.RequestQueryWithPolicy("actuator/allow_endpoint"))
		di.ActrReg.MustRegister(acCustomizer)
	}
	if di.HealthReg != nil {
		di.HealthReg.MustRegister(opaactuator.NewHealthDisclosureControlWithOPA(opa.QueryWithPolicy("actuator/allow_health_details")))
	}

	di.SecReg.Register(&securityConfigurer{})
}
