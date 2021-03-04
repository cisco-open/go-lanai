package servicedisc

const (
	applicationPropertiesPrefix = "application"
	discoveryPropertiesPrefix   = "cloud.consul.discovery"
)

type DiscoveryProperties struct {
	HealthCheckPath            string   `json:"health-check-path"`
	HealthCheckInterval        string   `json:"health-check-interval"`
	Tags                       []string `json:"tags"`
	AclToken                   string   `json:"acl-token"`
	IpAddress                  string   `json:"ip-address"` //A pre-defined IP address
	Interface                  string   `json:"interface"`    //The network interface from where to get the ip address. If IpAddress is defined, this field is ignored
	Port                       int      `json:"port"`
	Scheme                     string   `json:"scheme"`
	HealthCheckCriticalTimeout string   `json:"health-check-critical-timeout"` //See api.AgentServiceCheck's DeregisterCriticalServiceAfter field
	//TODO: add other values if needed
}
