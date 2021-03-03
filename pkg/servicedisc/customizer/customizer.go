package sdcustomizer

import (
	"errors"
	"github.com/hashicorp/consul/api"
)

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