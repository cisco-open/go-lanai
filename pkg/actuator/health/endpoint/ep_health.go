package health

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/json"
)

const (
	ID                   = "health"
	EnableByDefault      = true
)

type Input struct{}

type Output struct {
	health.Health
	sc int
}

type CompositeHealthV2 struct {
	health.SimpleHealth
	Components map[string]health.Health `json:"details,omitempty"`
}

// StatusCode http.StatusCoder
func (o Output) StatusCode() int {
	return o.sc
}

// Body web.BodyContainer
func (o Output) Body() interface{} {
	return o.Health
}

// MarshalJSON json.Marshaler
func (o Output) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Health)
}

type EndpointOptions func(opt *EndpointOption)
type EndpointOption struct {
	Contributor      health.Indicator
	StatusCodeMapper health.StatusCodeMapper
	MgtProperties    *actuator.ManagementProperties
	Properties       *health.HealthProperties
}

// HealthEndpoint implements actuator.Endpoint, actuator.WebEndpoint
type HealthEndpoint struct {
	actuator.WebEndpointBase
	contributor    health.Indicator
	scMapper       health.StatusCodeMapper
	showDetails    health.ShowMode
	showComponents health.ShowMode
	permissions    utils.StringSet
}

func newEndpoint(opts ...EndpointOptions) *HealthEndpoint {
	opt := EndpointOption{}
	for _, f := range opts {
		f(&opt)
	}

	if opt.StatusCodeMapper == nil {
		scMapper := health.DefaultStaticStatusCodeMapper
		for k, v := range opt.Properties.Status.ScMapping {
			scMapper[k] = v
		}
		opt.StatusCodeMapper = scMapper
	}

	showComponents := opt.Properties.ShowDetails
	if opt.Properties.ShowComponents != nil {
		showComponents = *opt.Properties.ShowComponents
	}
	ep := HealthEndpoint{
		contributor:    opt.Contributor,
		scMapper:       opt.StatusCodeMapper,
		showDetails:    opt.Properties.ShowDetails,
		showComponents: showComponents,
		permissions:    utils.NewStringSet(opt.Properties.Permissions...),
	}

	properties := opt.MgtProperties
	ep.WebEndpointBase = actuator.MakeWebEndpointBase(func(opt *actuator.EndpointOption) {
		opt.Id = ID
		opt.Ops = []actuator.Operation{
			actuator.NewReadOperation(ep.Read),
		}
		opt.Properties = &properties.Endpoints
		opt.EnabledByDefault = EnableByDefault
	})
	return &ep
}

// Read never returns error
func (ep *HealthEndpoint) Read(ctx context.Context, _ *Input) (*Output, error) {
	opts := health.Options{
		ShowDetails:    ep.shouldShowDetails(ctx),
		ShowComponents: ep.shouldShowComponents(ctx),
	}
	h := ep.contributor.Health(ctx, opts)
	switch f := ep.WebEndpointBase.NegotiateFormat(ctx); f {
	case actuator.ContentTypeSpringBootV2:
		h = ep.toSpringBootV2(h)
	}

	// Note: we know that *SystemHealthIndicator respect options (as all CompositeIndicator)
	// we don't need to sanitize result
	return &Output{
		Health: h,
		sc:     ep.scMapper.StatusCode(ctx, h.Status()),
	}, nil
}

func (ep *HealthEndpoint) isAuthorized(ctx context.Context) bool {
	auth := security.Get(ctx)
	if auth.State() < security.StateAuthenticated || auth.Permissions() == nil {
		return false
	}
	for p, _ := range ep.permissions {
		if _, ok := auth.Permissions()[p]; !ok {
			return false
		}
	}

	return true
}

func (ep *HealthEndpoint) shouldShowDetails(ctx context.Context) bool {
	switch ep.showDetails {
	case health.ShowModeNever:
		return false
	case health.ShowModeAlways:
		return true
	default:
		return ep.isAuthorized(ctx)
	}
}

func (ep *HealthEndpoint) shouldShowComponents(ctx context.Context) bool {
	switch ep.showComponents {
	case health.ShowModeNever:
		return false
	case health.ShowModeAlways:
		return true
	default:
		return ep.isAuthorized(ctx)
	}
}

func (ep *HealthEndpoint) toSpringBootV2(h health.Health) health.Health {
	var composite *health.CompositeHealth
	switch v := h.(type) {
	case health.CompositeHealth:
		composite = &v
	case *health.CompositeHealth:
		composite = v
	default:
		return h
	}
	
	ret := CompositeHealthV2{
		SimpleHealth: composite.SimpleHealth,
		Components:   make(map[string]health.Health),
	}
	// recursively convert components
	for k, v := range composite.Components {
		ret.Components[k] = ep.toSpringBootV2(v)
	}
	return ret
}
