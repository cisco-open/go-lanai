package opaactuator

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"github.com/open-policy-agent/opa/sdk"
)

func NewHealthDisclosureControlWithOPA(opaEngine *sdk.OPA, policy string) health.DisclosureControl {
	return health.DisclosureControlFunc(func(ctx context.Context) bool {
		e := opa.Allow(ctx, func(q *opa.Query) {
			q.OPA = opaEngine
			q.Policy = policy
		})
		return e == nil
	})

}
