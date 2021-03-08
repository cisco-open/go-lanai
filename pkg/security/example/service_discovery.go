package example

import (
	"context"
	"github.com/hashicorp/consul/api"
)

type RegistrationCustomizer struct {

}

func (r *RegistrationCustomizer) Customize(ctx context.Context, registration *api.AgentServiceRegistration) {
	registration.Tags = append(registration.Tags, "a=b")
}