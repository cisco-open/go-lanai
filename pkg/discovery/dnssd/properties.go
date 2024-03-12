package dnssd

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const (
	PropertiesPrefix = "cloud.discovery.dns"
)

// DiscoveryProperties defines static configuration of DNS SRV lookup
// SRV lookup is in format of "_<service>._<proto>.<target>" or "<target>"
// e.g. _http._tcp.my-service.my-namespace.svc.cluster.local
//      my-service.my-namespace.svc.cluster.local
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
	SRVProto          string `json:"srv-proto"`

	// SRVService is the "Service" defined in RFC2782 (The symbolic name of the desired service).
	// When present, the value need to be prepended with underscore "_", And depending on the deployment environment,
	// this could have different values.
	// e.g. Kubernetes define this value to be the "port name", and Consul doesn't support "Proto" and "Service" in static DNS queries
	// Optional, when specified, SRVProto should also be specified
	SRVService        string `json:"srv-service"`

	// FQDNFallback whether to use FQDN as-is, in case that SRV lookup yields no result.
	// When enabled, service discovery would never fail:
	// If SRV lookup fails, a single instance with no tags/meta and zeroed port number is returned
	FQDNFallback bool `json:"fqdn-fallback"`
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
