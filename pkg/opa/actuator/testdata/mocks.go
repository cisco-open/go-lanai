package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
)

const SpecialScopeAdmin = `admin`

type MockedHealthIndicator struct {
	Status      health.Status
	Description string
	Details     map[string]interface{}
}

func NewMockedHealthIndicator() *MockedHealthIndicator {
	return &MockedHealthIndicator{
		Status: health.StatusUp,
		Description: "mocked",
		Details:     map[string]interface{}{
			"key": "value",
		},
	}
}

func (i *MockedHealthIndicator) Name() string {
	return "test"
}

func (i *MockedHealthIndicator) Health(_ context.Context, opts health.Options) health.Health {
	ret := health.CompositeHealth{
		SimpleHealth: health.SimpleHealth{
			Stat: i.Status,
			Desc: i.Description,
		},
	}
	if opts.ShowComponents {
		detailed := health.DetailedHealth{
			SimpleHealth: health.SimpleHealth{
				Stat: i.Status,
				Desc: "mocked detailed",
			},
		}
		if opts.ShowDetails {
			detailed.Details = i.Details
		}

		ret.Components = map[string]health.Health{
			"simple": health.SimpleHealth{
				Stat: i.Status,
				Desc: "mocked simple",
			},
			"detailed": detailed,
		}
	}
	return ret
}
