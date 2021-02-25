package health

import (
	"context"
	"math"
)

/*******************************
	SimpleStatusAggregator
********************************/
var (
	DefaultStatusOrders = []Status{
		StatusDown, StatusOutOfService, StatusUp, StatusDown,
	}
)

type AggregateOptions func(opt *AggregateOption)
type AggregateOption struct {
	StatusOrders []Status
}

// SimpleStatusAggregator implements StatusAggregator
type SimpleStatusAggregator struct {
	orders map[Status]int
}

func NewSimpleStatusAggregator(opts ...AggregateOptions) *SimpleStatusAggregator {
	opt := AggregateOption{
		StatusOrders: DefaultStatusOrders,
	}
	for _, f := range opts {
		f(&opt)
	}
	orders := map[Status]int{}
	for i, s := range opt.StatusOrders {
		orders[s] = i
	}
	return &SimpleStatusAggregator{
		orders: orders,
	}
}

func (a SimpleStatusAggregator) Aggregate(_ context.Context, statuses ...Status) Status {
	var status *Status = nil
	for _, s := range statuses {
		if status == nil || a.compare(s, *status) < 0 {
			status = &s
		}
	}

	if status == nil {
		return StatusUnkown
	}
	return *status
}

func (a SimpleStatusAggregator) compare(s1, s2 Status) int {
	o1, ok := a.orders[s1]
	if !ok {
		o1 = math.MaxInt64
	}

	o2, ok := a.orders[s2]
	if !ok {
		o2 = math.MaxInt64
	}

	switch {
	case o1 < o2:
		return -1
	case o1 > o2:
		return 1
	default:
		return 0
	}
}
