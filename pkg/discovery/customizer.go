package discovery

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"github.com/hashicorp/consul/api"
)

/*
	TODO: For reference the java IDM implementation have the following consul registration customizers
	1. swaggerPathConsulRegistrationCustomizer
	add tag: swaggerPath=/swagger
	2. SecurityCompatibilityRegistrationCustomizer
	add tag: SMCR=3901
	add metadata: SMCR:3901
	3. Build info in both tag and metadata
	TAG_SERVICE_NAME, "build.name",
	TAG_VERSION, "build.version",
	TAG_BUILD_DATE_TIME, "build.time",
	TAG_BUILD_NUMBER, "build.number"
 */


type Customizer interface {
	Customize(ctx context.Context, reg *api.AgentServiceRegistration)
}

type Customizers struct {
	Customizers utils.Set
	applied bool
}

func NewCustomizers(ctx *bootstrap.ApplicationContext) *Customizers {
	return &Customizers{
		Customizers: utils.NewSet(NewDefaultCustomizer(ctx)),
	}
}

func (r *Customizers) Add(c Customizer) {
	if r.applied {
		panic(errors.New("cannot add consul registration customizer because other customization has already been applied"))
	}
	r.Customizers.Add(c)
}

func (r *Customizers) Apply(ctx context.Context, registration *api.AgentServiceRegistration) {
	if r.applied {
		return
	}
	defer func() {
		r.applied = true
	}()

	for c, _ := range r.Customizers {
		c.(Customizer).Customize(ctx, registration)
	}
}