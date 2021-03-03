package servicedisc


const (
	applicationPropertiesPrefix = "application"
	discoveryPropertiesPrefix = "cloud.consul.discovery"
)

type ApplicationProperties struct {
	Name string `json:"name"`
}

type DiscoveryProperties struct {
	HealthCheckPath string `json:"health-check-path"`
	HealthCheckInterval string `json:"health-check-interval"`
	Tags	[]string `json:"tags"`
	AclToken string `json:"acl-token"`
	IpAddress string `json:"ip-address"`
	Port int `json:"port"`
	Scheme string `json:"scheme"`
	HealthCheckCriticalTimeout string `json:"health-check-critical-timeout"` //See api.AgentServiceCheck's DeregisterCriticalServiceAfter field
	//TODO: add other values if needed
}

