package security

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
)

var CompatibilityDiscoveryCustomizer = compatibilityDiscoveryCustomizer{}

// compatibilityDiscoveryCustomizer implements discovery.Customizer
type compatibilityDiscoveryCustomizer struct {}

func (c compatibilityDiscoveryCustomizer) Customize(ctx context.Context, reg *api.AgentServiceRegistration) {
	tag := fmt.Sprintf("%s=%s", CompatibilityReferenceTag, CompatibilityReference)
	reg.Tags = append(reg.Tags, tag)
	if reg.Meta == nil {
		reg.Meta = map[string]string{}
	}
	reg.Meta[CompatibilityReferenceTag] = CompatibilityReference
}

