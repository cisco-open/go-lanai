package actuator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	matcherutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"fmt"
	"net/http"
)

// ActuatorSecurityCustomizer is a single ActuatorSecurityCustomizer can be registered via Registrar
// ActuatorSecurityCustomizer is typically responsible to setup authentication scheme
// it should not configure access control, which is configured via properties
//goland:noinspection GoNameStartsWithPackageName
type ActuatorSecurityCustomizer interface {
	Configure(ws security.WebSecurity)
}

type DefaultActuatorSecurityCustomizer struct {

}

func (c *DefaultActuatorSecurityCustomizer) Configure(ws security.WebSecurity) {
	ws.With(tokenauth.New())
}

// actuatorSecurityConfigurer implements security.Configurer
type actuatorSecurityConfigurer struct {
	properties *ManagementProperties
	endpoints  WebEndpoints
	customizer ActuatorSecurityCustomizer
}

func (c *actuatorSecurityConfigurer) Configure(ws security.WebSecurity) {
	if c.customizer != nil {
		c.customizer.Configure(ws)
	}

	path := fmt.Sprintf("%s/**", c.properties.Endpoints.Web.BasePath)


	ws.Route(matcher.RouteWithPattern(path).And(matcherutils.Not(matcher.RouteWithMethods(http.MethodOptions)))).
		With(errorhandling.New())

	// configure access control based on properties and installed web endpoints
	ac := access.Configure(ws)
	for k, _ := range c.endpoints {
		c.configureAccessControl(ac, k, c.endpoints.Paths(k))
	}

	// fallback configuration
	if c.properties.Security.EnabledByDefault {
		ac.Request(matcher.AnyRequest()).Authenticated()
	} else {
		ac.Request(matcher.AnyRequest()).PermitAll()
	}
}

func (c *actuatorSecurityConfigurer) configureAccessControl(ac *access.AccessControlFeature, epId string, paths []string){
	if len(paths) == 0 {
		return
	}

	// first grab some facts
	enabled := c.properties.Security.EnabledByDefault
	permissions := c.properties.Security.Permissions
	if props, ok := c.properties.Security.Endpoints[epId]; ok {
		permissions = props.Permissions
		if props.Enabled != nil {
			enabled = *props.Enabled
		}
	}

	// second configure request matchers
	m := matcher.RequestWithPattern(paths[0])
	for _, p := range paths[1:] {
		m = m.Or(matcher.RequestWithPattern(p))
	}

	// third configure access control
	switch {
	case !enabled:
		ac.Request(m).PermitAll()
	case len(permissions) == 0:
		ac.Request(m).Authenticated()
	default:
		ac.Request(m).HasPermissions(permissions...)
	}
}
