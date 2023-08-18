package opaactuator

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
)

func NewHealthDisclosureControlWithOPA(opts ...opa.QueryOptions) health.DisclosureControl {
	return health.DisclosureControlFunc(func(ctx context.Context) bool {
		e := opa.Allow(ctx, opts...)
		return e == nil
	})

}
