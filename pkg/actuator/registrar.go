package actuator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"net/http"
)

type constructDI struct {
	fx.In
	Properties ManagementProperties
}

type initDI struct {
	fx.In
	ApplicationContext *bootstrap.ApplicationContext
	WebRegistrar       *web.Registrar     `optional:"true"`
	SecurityRegistrar  security.Registrar `optional:"true"`
}

type Registrar struct {
	initialized        bool
	properties         *ManagementProperties
	endpoints          []Endpoint
	securityConfigurer security.Configurer
	securityCustomizer ActuatorSecurityCustomizer
}

func NewRegistrar(di constructDI) *Registrar {
	return &Registrar{
		properties: &di.Properties,
		securityCustomizer: &DefaultActuatorSecurityCustomizer{},
	}
}

func (r *Registrar) initialize(di initDI) error {
	if r.initialized {
		return fmt.Errorf("attempting to initialize actuator twice")
	}

	defer func() {
		r.initialized = true
	}()

	// install web endpoints
	webEndpoints, e := r.installWebEndpoints(di.WebRegistrar)
	if e != nil {
		return e
	}
	logger.WithContext(di.ApplicationContext).
		Info(fmt.Sprintf("registered web endponts %v", webEndpoints.EndpointIDs()))

	// install security
	if e := r.installWebSecurity(di.SecurityRegistrar, webEndpoints); e != nil {
		return e
	}
	return nil
}

func (r *Registrar) MustRegister(items...interface{}) {
	if e := r.Register(items...); e != nil {
		panic(e)
	}
}

func (r *Registrar) Register(items...interface{}) error {
	for _, item := range items {
		if e := r.register(item); e != nil {
			return e
		}
	}
	return nil
}

func (r *Registrar) register(item interface{}) error {
	if r.initialized {
		return fmt.Errorf("attempting to register actuator items after actuator has been initialized")
	}

	switch item.(type) {
	case Endpoint:
		r.endpoints = append(r.endpoints, item.(Endpoint))
	case []interface{}:
		r.Register(item.([]interface{})...)
	case ActuatorSecurityCustomizer:
		r.securityCustomizer = item.(ActuatorSecurityCustomizer)
	default:
		return fmt.Errorf("unsupported actuator type [%T]", item)
	}
	return nil
}

func (r *Registrar) installWebEndpoints(reg *web.Registrar) (WebEndpoints, error) {
	if reg == nil || !r.properties.Enabled {
		return nil, nil
	}

	result := WebEndpoints{}
	for _, ep := range r.endpoints {
		if mappings, e := r.installWebEndpoint(reg, ep); e != nil {
			return nil, e
		} else if len(mappings) != 0 {
			result[ep.Id()] = mappings
		}
	}
	return result, nil
}

func (r *Registrar) installWebEndpoint(reg *web.Registrar, endpoint Endpoint) ([]web.Mapping, error) {

	if reg == nil || !r.isEndpointEnabled(endpoint) || !r.shouldExposeToWeb(endpoint) {
		return nil, nil
	}

	ops := endpoint.Operations()
	mappings := make([]web.Mapping, 0, len(ops))
	paths := utils.NewStringSet()
	for _, op := range ops {
		opMappings, e := endpoint.(WebEndpoint).Mappings(op, "")
		if e != nil {
			return nil, e
		}

		if e := reg.Register(opMappings); e != nil {
			return nil, e
		}
		mappings = append(mappings, opMappings...)
		for _, m := range opMappings {
			if route, ok := m.(web.RoutedMapping); ok {
				paths.Add(route.Group() + route.Path())
			}
		}
	}
	// add OPTIONS route
	for path := range paths {
		m := mapping.Options(path).HandlerFunc(optionsHttpHandlerFunc()).Build()
		if e := reg.Register(m); e != nil {
			return nil, e
		}
	}
	return mappings, nil
}

func (r *Registrar) installWebSecurity(reg security.Registrar, endpoints WebEndpoints) error {
	if reg == nil {
		return nil
	}

	configurer := actuatorSecurityConfigurer{
		properties: r.properties,
		endpoints:  endpoints,
		customizer: r.securityCustomizer,
	}
	reg.Register(&configurer)

	return nil
}

/*******************************
	internal
********************************/
func (r *Registrar) isEndpointEnabled(endpoint Endpoint) bool {
	if !r.properties.Enabled {
		return false
	}

	if basic, ok := r.properties.BasicEndpoint[endpoint.Id()]; !ok || basic.Enabled == nil {
		// not explicitly specified, use default
		return endpoint.EnabledByDefault() || r.properties.Endpoints.EnabledByDefault
	} else {
		return *basic.Enabled
	}
}

func (r *Registrar) shouldExposeToWeb(endpoint Endpoint) bool {
	if _, ok := endpoint.(WebEndpoint); !ok {
		return false
	}

	includeAll := r.properties.Endpoints.Web.Exposure.Include.Has("*")
	include := r.properties.Endpoints.Web.Exposure.Include.Has(endpoint.Id())
	excludeAll := r.properties.Endpoints.Web.Exposure.Exclude.Has("*")
	exclude := r.properties.Endpoints.Web.Exposure.Include.Has(endpoint.Id())
	switch {
	case !excludeAll && !exclude && (includeAll || include):
		// no exclusion & include is set
		return true
	case !exclude && include:
		// no explicit exclusion & explicit inclusion
		return true
	default:
		// explicit exclusion or implicit exclusion without explicit inclusion
		return false
	}
}


func optionsHttpHandlerFunc() gin.HandlerFunc {
	return func(gc *gin.Context) {
		gc.Status(http.StatusOK)
	}
}