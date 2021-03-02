package servicedisc


const (
	applicationPropertiesPrefix = "application"
	discoveryPropertiesPrefix = "cloud.consul.discovery"
)

type ApplicationProperties struct {
	Name string `json:"name"`
}

/*
spring.cloud.consul.discovery.instance-id=${spring.application.name}-${local.server.port:${server.port:0}}-${spring.cloud.consul.discovery.bootstrap-id}
spring.cloud.consul.discovery.fail-fast=true
spring.cloud.consul.discovery.health-check-path=${server.servlet.context-path:}${management.endpoints.web.base-path:}/health
spring.cloud.consul.discovery.health-check-interval=15s
spring.cloud.consul.discovery.port=${management.server.port:${local.server.port:${server.port:0}}}
 */
type DiscoveryProperties struct {
	FailFast bool `json:"fail-fast"`
	HealthCheckPath string `json:"health-check-path"`
	HealthCheckInterval string `json:"health-check-interval"`
	Tags	[]string `json:"tags"`
	AclToken string `json:"acl-token"`
	IpAddress string `json:"ip-address"`
	Port int `json:"port"`
	Scheme string `json:"scheme"`
	HealthCheckCriticalTimeout string `json:"health-check-critical-timeout"`
	//TODO: add other values if needed
}

