package bootstrap

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	"strings"
)

const (
	TAG_VERSION = "version";
	TAG_BUILD_NUMBER = "buildNumber";
	TAG_BUILD_DATE_TIME = "buildDateTime";
)

var BuildInfoDiscoveryCustomizer = buildInfoDiscoveryCustomizer{}

type buildInfoDiscoveryCustomizer struct {}

func (b buildInfoDiscoveryCustomizer) Customize(ctx context.Context, reg *api.AgentServiceRegistration) {
	attrs := map[string]string {
		TAG_VERSION: BuildVersion,
		TAG_BUILD_DATE_TIME: BuildTime,
	}

	components := strings.Split(BuildVersion, "-")
	if len(components) == 2 {
		attrs[TAG_BUILD_NUMBER] = components[1]
	}

	if reg.Meta == nil {
		reg.Meta = map[string]string{}
	}

	for k, v := range attrs {
		reg.Meta[k] = v
		reg.Tags = append(reg.Tags, fmt.Sprintf("%s=%s", k, v))
	}
}