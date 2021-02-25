package health

import (
	"fmt"
)

func initialize(di regDI) error {

	endpoint := newEndpoint(func(opt *EndpointOption) {
		opt.MgtProperties = &di.MgtProperties
		opt.Contributor = di.HealthIndicator
		opt.Properties = &di.Properties
	})
	di.Registrar.Register(endpoint)
	return nil
}

type Registrar interface {
	// Register configure SystemHealthIndicator and HealthEndpoint
	// supported input parameters are:
	// 	- Indicator
	// 	- StatusAggregator
	Register(items ...interface{}) error
}

// SystemHealthIndicator implements Indicator and Registrar
type SystemHealthIndicator struct {
	CompositeIndicator
}

func newSystemHealthIndicator(di provideDI) *SystemHealthIndicator {
	return &SystemHealthIndicator{
		CompositeIndicator: CompositeIndicator{
			name:       "system",
			delegates:  []Indicator{
				&PingIndicator{},
			},
			aggregator: NewSimpleStatusAggregator(func(opt *AggregateOption) {
				if len(di.Properties.Status.Orders) != 0 {
					opt.StatusOrders = di.Properties.Status.Orders
				}
			}),
		},
	}
}

// Register configure SystemHealthIndicator
// supported input parameters are:
// 	- Indicator
// 	- StatusAggregator
func (i *SystemHealthIndicator) Register(items ...interface{}) error {
	for _, v := range items {
		if e := i.register(v); e != nil {
			return e
		}
	}
	return nil
}

func (i *SystemHealthIndicator) register(v interface{}) error {
	switch v.(type) {
	case []interface{}:
		return i.Register(v.([]interface{})...)
	case Indicator:
		i.Add(v.(Indicator))
	case StatusAggregator:
		i.aggregator = v.(StatusAggregator)
	default:
		return fmt.Errorf("unsupported item %T", v)
	}
	return nil
}

