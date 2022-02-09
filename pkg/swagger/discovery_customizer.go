package swagger

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
)

const (
	TAG_SWAGGER_PATH = "swaggerPath"
)

type swaggerInfoDiscoveryCustomizer struct {}

func (s swaggerInfoDiscoveryCustomizer) Customize(ctx context.Context, reg *api.AgentServiceRegistration) {
	reg.Tags = append(reg.Tags, fmt.Sprintf("%s=%s", TAG_SWAGGER_PATH, "/swagger"))
}
