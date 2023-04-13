package health

import (
	"fmt"
	"go.uber.org/fx"
)

// SystemHealthRegistrar implements Registrar
type SystemHealthRegistrar struct {
	Indicator            *CompositeIndicator
	DetailsDisclosure    DetailsDisclosureControl
	ComponentsDisclosure ComponentsDisclosureControl
}

type regDI struct {
	fx.In
	Properties    HealthProperties
}

func NewSystemHealthRegistrar(di regDI) *SystemHealthRegistrar {
	return &SystemHealthRegistrar{
		Indicator: &CompositeIndicator{
			name: "system",
			delegates: []Indicator{
				PingIndicator{},
			},
			aggregator: NewSimpleStatusAggregator(func(opt *AggregateOption) {
				if len(di.Properties.Status.Orders) != 0 {
					opt.StatusOrders = di.Properties.Status.Orders
				}
			}),
		},
	}
}

// Register configure SystemHealthRegistrar
// supported input parameters are:
// 	- Indicator
// 	- StatusAggregator
func (i *SystemHealthRegistrar) Register(items ...interface{}) error {
	for _, v := range items {
		if e := i.register(v); e != nil {
			return e
		}
	}
	return nil
}

func (i *SystemHealthRegistrar) MustRegister(items ...interface{}) {
	if e := i.Register(items...); e != nil {
		panic(e)
	}
}

func (i *SystemHealthRegistrar) register(item interface{}) error {
	switch v := item.(type) {
	case []interface{}:
		return i.Register(v...)
	case Indicator:
		i.Indicator.Add(v)
	case StatusAggregator:
		i.Indicator.aggregator = v
	case DisclosureControl:
		i.DetailsDisclosure = v
		i.ComponentsDisclosure = v
	case DetailsDisclosureControl:
		i.DetailsDisclosure = v
	case ComponentsDisclosureControl:
		i.ComponentsDisclosure = v
	default:
		return fmt.Errorf("unsupported item %T", item)
	}
	return nil
}
