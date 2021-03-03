package sdcustomizer

import (
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
	Customize(registration *api.AgentServiceRegistration)
}

type Registrar struct {
	Customizers []Customizer
	applied bool
}

func NewRegistrar() *Registrar {
	return &Registrar{}
}

func (r *Registrar) Add(c Customizer) {
	if r.applied {
		panic(errors.New("cannot add consul registration customizer because other customization has already been applied"))
	}
	r.Customizers = append(r.Customizers, c)
}

func (r *Registrar) Apply(registration *api.AgentServiceRegistration) {
	for _, c := range r.Customizers {
		c.Customize(registration)
	}
	r.applied = true
}