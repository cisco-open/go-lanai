package discovery

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/pkg/errors"
)

//goland:noinspection GoNameStartsWithPackageName
const (
	applicationPropertiesPrefix = "application"
	DiscoveryPropertiesPrefix   = "cloud.consul.discovery"
)

//goland:noinspection GoNameStartsWithPackageName
type DiscoveryProperties struct {
	HealthCheckPath            string                    `json:"health-check-path"`
	HealthCheckInterval        string                    `json:"health-check-interval"`
	Tags                       utils.CommaSeparatedSlice `json:"tags"`
	AclToken                   string                    `json:"acl-token"`
	IpAddress                  string                    `json:"ip-address"` //A pre-defined IP address
	Interface                  string                    `json:"interface"`  //The network interface from where to get the ip address. If IpAddress is defined, this field is ignored
	Port                       int                       `json:"port"`
	Scheme                     string                    `json:"scheme"`
	HealthCheckCriticalTimeout string                    `json:"health-check-critical-timeout"` //See api.AgentServiceCheck's DeregisterCriticalServiceAfter field
}

func NewDiscoveryProperties() *DiscoveryProperties {
	return &DiscoveryProperties{
		Port:                0,
		Scheme:              "http",
		HealthCheckInterval: "15s",
		HealthCheckPath:     fmt.Sprintf("%s", "/admin/health"),
	}
}

func BindDiscoveryProperties(ctx *bootstrap.ApplicationContext) DiscoveryProperties {
	props := NewDiscoveryProperties()
	if err := ctx.Config().Bind(props, DiscoveryPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind DiscoveryProperties"))
	}
	return *props
}
