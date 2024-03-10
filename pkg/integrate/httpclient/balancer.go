package httpclient

import (
	"fmt"
	"sync/atomic"
)

type balancer[T any] interface {
	Balance([]T) (T, error)
}

func newRoundRobinBalancer[T any]() balancer[T] {
	return &roundRobin[T]{}
}

type roundRobin[T any] struct {
	c uint64
}

func (rr *roundRobin[T]) Balance(values []T) (v T, err error) {
	if len(values) <= 0 {
		return v, NewNoEndpointFoundError(fmt.Errorf("cannot find service"))
	}
	old := atomic.AddUint64(&rr.c, 1) - 1
	idx := old % uint64(len(values))
	return values[idx], nil
}
