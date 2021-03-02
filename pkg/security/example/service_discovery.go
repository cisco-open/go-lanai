package example

import (
	"github.com/hashicorp/consul/api"
)

type RegistrationCusomizer struct {

}

func (r *RegistrationCusomizer) Customize(registration *api.AgentServiceRegistration) {
	registration.Tags = append(registration.Tags, "a=b")
}