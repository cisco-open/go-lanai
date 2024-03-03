package discovery

import (
	"context"
)

// ServiceRegistration is the data to be registered with any external service registration system.
// It contains information about current running service instance.
// The implementation depends on which service discovery tech-stack is used.
// e.g. Consul would be *consulsd.ServiceRegistration
type ServiceRegistration interface {
	ID() string
	Name() string
	Address() string
	Port() int
	Tags() []string
	Meta() map[string]any

	SetID(id string)
	SetName(name string)
	SetAddress(addr string)
	SetPort(port int)
	AddTags(tags...string)
	RemoveTags(tags...string)
	SetMeta(key string, value any)
}

// ServiceRegistrar is the interface to interact with external service registration system.
type ServiceRegistrar interface {
	Register(ctx context.Context, registration ServiceRegistration) error
	Deregister(ctx context.Context, registration ServiceRegistration) error
}

// ServiceRegistrationCustomizer customize given ServiceRegistration during bootstrap.
// Any ServiceRegistrationCustomizer provided with fx group defined as FxGroup will be applied automatically.
type ServiceRegistrationCustomizer interface {
	Customize(ctx context.Context, reg ServiceRegistration)
}

// ServiceRegistrationCustomizerFunc is the func that implements ServiceRegistrationCustomizer
type ServiceRegistrationCustomizerFunc func(ctx context.Context, reg ServiceRegistration)

func (fn ServiceRegistrationCustomizerFunc) Customize(ctx context.Context, reg ServiceRegistration) {
	fn(ctx, reg)
}


