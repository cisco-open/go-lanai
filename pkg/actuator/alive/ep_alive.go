package alive

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"net/http"
)

const (
	ID                   = "alive"
	EnableByDefault      = true
)

type Input struct{}

type Output struct{
	sc int
	Message string `json:"msg"`
}

// http.StatusCoder
func (o Output) StatusCode() int {
	return o.sc
}

// AliveEndpoint implements actuator.Endpoint, actuator.WebEndpoint
type AliveEndpoint struct {
	actuator.WebEndpointBase
}

func New(di regDI) *AliveEndpoint {
	ep := AliveEndpoint{}
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
func (ep *AliveEndpoint) Read(ctx context.Context, input *Input) (Output, error) {
	return Output{
		sc: http.StatusOK,
		Message: "I'm good",
	}, nil
}
