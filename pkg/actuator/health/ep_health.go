package health

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"net/http"
)

const (
	ID                   = "health"
	EnableByDefault      = true
)

type Input struct{}

type Health struct{
	sc     int
	Status string `json:"status"`
}

// http.StatusCoder
func (o Health) StatusCode() int {
	return o.sc
}

// HealthEndpoint implements actuator.Endpoint, actuator.WebEndpoint
type HealthEndpoint struct {
	actuator.WebEndpointBase
}

func new(di regDI) *HealthEndpoint {
	ep := HealthEndpoint{}
	ep.WebEndpointBase = actuator.MakeWebEndpointBase(func(opt *actuator.EndpointOption) {
		opt.Id = ID
		opt.Ops = []actuator.Operation{
			actuator.NewReadOperation(ep.Read),
		}
		opt.Properties = &di.MgtProperties.Endpoints
		opt.EnabledByDefault = EnableByDefault
	})
	return &ep
}

// Read never returns error
func (ep *HealthEndpoint) Read(ctx context.Context, input *Input) (*Health, error) {
	return &Health{
		sc: http.StatusOK,
		Status: StatusUp,
	}, nil
}
