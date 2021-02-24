package info

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
)

const (
	ID                   = "info"
	EnableByDefault      = true
	infoPropertiesPrefix = "info"
)

type Input struct {
	Name string `uri:name`
}

type Info map[string]interface{}

// InfoEndpoint implements actuator.Endpoint, actuator.WebEndpoint
type InfoEndpoint struct {
	actuator.WebEndpointBase
	appConfig appconfig.ConfigAccessor
}

func New(di regDI) *InfoEndpoint {
	ep := InfoEndpoint{
		appConfig: di.AppContext.Config(),
	}
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
func (ep *InfoEndpoint) Read(ctx context.Context, input *Input) (interface{}, error) {
	info := Info{}
	if e := ep.appConfig.Bind(&info, infoPropertiesPrefix); e != nil {
		return Info{}, nil
	}

	if input.Name == "" {
		return info, nil
	}

	if v, ok := info[input.Name]; ok {
		return v, nil
	}
	return Info{}, nil
}


