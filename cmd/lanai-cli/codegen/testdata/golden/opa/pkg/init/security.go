package serviceinit

import (
	"github.com/cisco-open/go-lanai/pkg/actuator"
	"github.com/cisco-open/go-lanai/pkg/actuator/health"
	"github.com/cisco-open/go-lanai/pkg/opa"
	opaaccess "github.com/cisco-open/go-lanai/pkg/opa/access"
	opaactuator "github.com/cisco-open/go-lanai/pkg/opa/actuator"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/access"
	"github.com/cisco-open/go-lanai/pkg/security/config/resserver"
	"github.com/cisco-open/go-lanai/pkg/security/errorhandling"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/tokenauth"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
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
			CustomDecisionMaker(opaaccess.DecisionMakerWithOPA(opa.RequestQueryWithPolicy("testservice_api/allow_api"))),
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
		di.HealthReg.MustRegister(opaactuator.NewHealthDisclosureControlWithOPA(
			opa.QueryWithPolicy("actuator/allow_health_details"),
			opa.SilentQuery(),
		))
	}

	di.SecReg.Register(&securityConfigurer{})
}
