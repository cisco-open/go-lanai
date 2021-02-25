package actuator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"go.uber.org/fx"
)

type constructDI struct {
	fx.In
	Properties ManagementProperties
}

type initDI struct {
	fx.In
	WebRegistrar      *web.Registrar `optional:"true"`
	SecurityRegistrar security.Registrar `optional:"true"`
}

type Registrar struct {
	initialized bool
	properties  ManagementProperties
	endpoints   []Endpoint
}

func NewRegistrar(di constructDI) *Registrar {
	return &Registrar{
		properties: di.Properties,
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
	webEndpoints := []string{}
	for _, ep := range r.endpoints {
		if isWeb, e := r.installWebEndpoint(di.WebRegistrar, ep); e != nil {
			return e
		} else if isWeb {
			webEndpoints = append(webEndpoints, ep.Id())
		}
	}
	logger.Info(fmt.Sprintf("registered web endponts %v", webEndpoints))

	return nil
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
	default:
		return fmt.Errorf("unsupported actuator type [%T]", item)
	}
	return nil
}

func (r *Registrar) installWebEndpoint(reg *web.Registrar, endpoint Endpoint) (bool, error) {

	if reg == nil || !r.isEndpointEnabled(endpoint) || !r.shouldExposeToWeb(endpoint) {
		return false, nil
	}

	for _, op := range endpoint.Operations() {
		mapping, e := endpoint.(WebEndpoint).Mapping(op, "")
		if e != nil {
			return true, e
		}

		if e := reg.Register(mapping); e != nil {
			return true, e
		}
	}
	return true, nil
}

/*******************************
	internal
********************************/
func (r *Registrar) isEndpointEnabled(endpoint Endpoint) bool {
	if !r.properties.Enabled {
		return false
	}

	if basic, ok := r.properties.BasicEndpoint[endpoint.Id()]; !ok {
		return endpoint.EnabledByDefault() || r.properties.Endpoints.EnabledByDefault
	} else {
		return basic.Enabled
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
