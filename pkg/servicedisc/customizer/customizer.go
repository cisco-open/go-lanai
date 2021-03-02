package sdcustomizer

import "github.com/hashicorp/consul/api"

type ConsulRegistrationCustomizer interface {
	Customize(registration *api.AgentServiceRegistration)
}

var Customizers []ConsulRegistrationCustomizer
