package discovery

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"github.com/pkg/errors"
)

const (
	applicationPropertiesPrefix = "application"
	DiscoveryPropertiesPrefix   = "cloud.consul.discovery"
)

type DiscoveryProperties struct {
	HealthCheckPath            string   `json:"health-check-path"`
	HealthCheckInterval        string   `json:"health-check-interval"`
	Tags                       utils.CommaSeparatedSlice `json:"tags"`
	AclToken                   string   `json:"acl-token"`
	IpAddress                  string   `json:"ip-address"` //A pre-defined IP address
	Interface                  string   `json:"interface"`    //The network interface from where to get the ip address. If IpAddress is defined, this field is ignored
	Port                       int      `json:"port"`
	Scheme                     string   `json:"scheme"`
	HealthCheckCriticalTimeout string   `json:"health-check-critical-timeout"` //See api.AgentServiceCheck's DeregisterCriticalServiceAfter field
	//TODO: add other values if needed
}

func NewDiscoveryProperties(serverProps *web.ServerProperties) *DiscoveryProperties {
	return &DiscoveryProperties{
		Port: serverProps.Port,
		Scheme: "http",
		HealthCheckInterval: "15s",
		HealthCheckPath: fmt.Sprintf("%s%s", serverProps.ContextPath, "/admin/health"),
	}
}

func BindDiscoveryProperties(ctx *bootstrap.ApplicationContext, serverProps web.ServerProperties) DiscoveryProperties {
	props := NewDiscoveryProperties(&serverProps)
	if err := ctx.Config().Bind(props, DiscoveryPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind DiscoveryProperties"))
	}
	return *props
}
