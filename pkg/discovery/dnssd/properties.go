package dnssd

import (
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
	"regexp"
)

const (
	PropertiesPrefix = "cloud.discovery.dns"
)

// DiscoveryProperties defines static configuration of DNS SRV lookup
// SRV lookup is in format of "_<service>._<proto>.<target>" or "<target>"
// e.g. _http._tcp.my-service.my-namespace.svc.cluster.local
//
//	my-service.my-namespace.svc.cluster.local
//
// See [RFC2782](https://datatracker.ietf.org/doc/html/rfc2782)
type DiscoveryProperties struct {
	// Addr is the address of DNS server. e.g. "8.8.8.8:53"
	// If not set, default DNS server is used.
	// Note: Resolving DNS server address may also require DNS lookup. Please set this value with caution
	Addr string `json:"addr"`

	// FQDNTemplate is a golang template with single-line to define how to
	// translate service name to target domain name (<target>) in DNS lookup query.
	// The template data contains field ".ServiceName".
	// e.g. "{{.ServiceName}}.my-namespace.svc.cluster.local"
	FQDNTemplate string `json:"fqdn-template"`

	// SRVProto is the "Proto" defined in RFC2782 (The symbolic name of the desired protocol).
	// When present, the value need to be prepended with underscore "_".
	// e.g. "_tcp", "_udp"
	// Optional, when specified, SRVService should also be specified
	SRVProto string `json:"srv-proto"`

	// SRVService is the "Service" defined in RFC2782 (The symbolic name of the desired service).
	// When present, the value need to be prepended with underscore "_", And depending on the deployment environment,
	// this could have different values.
	// e.g. Kubernetes define this value to be the "port name", and Consul doesn't support "Proto" and "Service" in static DNS queries
	// Optional, when specified, SRVProto should also be specified
	SRVService string `json:"srv-service"`

	// Fallback defines how does service discovery behave in case of DNS lookup couldn't resolve any instances.
	// See FallbackProperties for more details
	Fallback FallbackProperties `json:"fallback"`
}

// FallbackProperties defines host rewrite as the last resort
// In case DNS lookup fails, discovery client would try following:
//  1. If the service name matches any entry in Mappings, its HostMappingProperties.Hosts is used as healthy instances list.
//     This is equivalent to static service discovery.
//  2. If step 1 yield no result, Default is used and the result would be a resolved service with single instance.
//     This is equivalent to use server-side load balancing.
//  3. If none of above yield valid result, the original DNS lookup error is recorded and the service is temporarily undiscoverable.
type FallbackProperties struct {
	// Mappings defines how to map the service names to hosts, based on service name patterns.
	// The keys of the map is literal and does not affect mapping behaviour
	// All entries in Mappings  are tried in an undefined order, so make sure they don't have overlapped patterns
	Mappings map[string]HostMappingProperties `json:"mappings"`

	// Default is a golang template with single-line output to rewrite any service name into host.
	// The template data contains field ".ServiceName".
	// 		e.g. "{{.ServiceName}}.default.svc.cluster.local:8443"
	// This value is used when the service name is not applicable to any entry in Mappings .
	Default string `json:"default"`
}

func (p FallbackProperties) CompileMappings() ([]HostMapping, error) {
	mappings := make([]HostMapping, 0, len(p.Mappings)+1)
	for _, m := range p.Mappings {
		regex, e := regexp.CompilePOSIX(m.Service)
		if e != nil {
			return nil, fmt.Errorf(`invalid service name pattern "%s": %v`, m.Service, e)
		}
		mappings = append(mappings, HostMapping{ServiceRegex: regex, Hosts: m.Hosts})
	}
	if len(p.Default) != 0 {
		mappings = append(mappings, HostMapping{ServiceRegex: regexp.MustCompilePOSIX(`.+`), Hosts: []string{p.Default}})
	}
	return mappings, nil
}

type HostMappingProperties struct {
	// Service the name of the service, support regular expression
	Service string `json:"service"`
	// Hosts is a list of known hosts. Each entry should be a golang template with single-line output.
	// The template data contains field ".ServiceName"
	// 		e.g. "pod-1.{{.ServiceName}}.default.svc.cluster.local:8989"
	Hosts []string `json:"hosts"`
}

func NewDiscoveryProperties() *DiscoveryProperties {
	return &DiscoveryProperties{
		FQDNTemplate: "{{.ServiceName}}.default.svc.cluster.local",
	}
}

func BindDiscoveryProperties(ctx *bootstrap.ApplicationContext) DiscoveryProperties {
	props := NewDiscoveryProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind DiscoveryProperties"))
	}
	return *props
}
